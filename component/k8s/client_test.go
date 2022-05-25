package k8s

import (
	"context"
	"testing"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme = runtime.NewScheme()
)

var (
	errKubeConfig = "errorKubeConfig"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

func TestNewClient(t *testing.T) {
	cli, err := NewClient(
		"test-cluster",
		WithRuntimeManagerOptions(manager.Options{
			Scheme: scheme,
		}),
	)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Stop()

	t.Run("unregistry scheme object", func(t *testing.T) {
		err := cli.AddEventHandler(&networkingv1beta1.DestinationRule{}, cache.ResourceEventHandlerFuncs{})
		if err == nil {
			t.Errorf("networkingv1beta1.DestinationRule not registry, must be error")
		}
	})

	t.Run("AddEventHandler pod", func(t *testing.T) {
		err := cli.AddEventHandler(&corev1.Pod{}, cache.ResourceEventHandlerFuncs{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("client not start", func(t *testing.T) {
		if cli.HasSynced() {
			t.Error("client not start must return false")
		}
	})
	t.Run("client get not start", func(t *testing.T) {
		err := cli.GetObj(types.NamespacedName{Namespace: "default", Name: "zk"}, &corev1.Pod{})
		if err == nil {
			t.Error("client not start must return false")
		}
	})

	t.Run("client create not start", func(t *testing.T) {
		err := cli.CreateObj(&corev1.Pod{})
		if err == nil {
			t.Error("client not start must return false")
		}
	})

	t.Run("client update not start", func(t *testing.T) {
		err := cli.UpdateObj(&corev1.Pod{})
		if err == nil {
			t.Error("client not start must return false")
		}
	})

	t.Run("client update status not start", func(t *testing.T) {
		err := cli.UpdateObjStatus(&corev1.Pod{})
		if err == nil {
			t.Error("client not start must return false")
		}
	})

	t.Run("client delete not start", func(t *testing.T) {
		err := cli.DeleteObj(&corev1.Pod{})
		if err == nil {
			t.Error("client not start must return false")
		}
	})

	go func() {
		if err := cli.Start(context.TODO()); err != nil {
			t.Log(err)
		}
	}()

	for !cli.HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}

	t.Run("client CRUD", func(t *testing.T) {
		// create
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "zk",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image:           "zookeeper",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "zk",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 2181,
								Protocol:      corev1.ProtocolTCP,
							},
						},
					},
				},
			},
		}
		err := cli.CreateObj(pod)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(time.Second * 5)

		// get
		p := &corev1.Pod{}
		err = cli.GetObj(types.NamespacedName{Namespace: "default", Name: "zk"}, p)
		if err != nil {
			t.Error(err)
			return
		}

		p.Labels = map[string]string{"aa": "bb"}
		err = cli.UpdateObj(p)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(time.Second * 5)

		err = cli.GetObj(types.NamespacedName{Namespace: "default", Name: "zk"}, p)
		if err != nil {
			t.Error(err)
			return
		}

		p.Status.Message = "test msg"
		err = cli.UpdateObjStatus(p)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(time.Second * 5)

		err = cli.UpdateObj(p)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(time.Second * 5)

		err = cli.DeleteObj(p)
		if err != nil {
			t.Error(err)
			return
		}
	})
}

func TestNewClientEventHandler(t *testing.T) {
	testNs := "default"

	cli, err := NewClient(
		"test-cluster",
		WithRuntimeManagerOptions(manager.Options{
			Scheme: scheme,
		}),
	)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Stop()

	pods, err := cli.KubeInterface.CoreV1().Pods(testNs).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Errorf("get default pods failed:%+v", err)
		return
	}
	podCount := len(pods.Items)

	cacheCount := 0
	cli.AddEventHandler(&corev1.Pod{}, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				t.Errorf("transfor failed:%+v", err)
				return
			}
			ns, _, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				t.Errorf("split namespace failed:%+v", err)
				return
			}
			if ns == testNs {
				cacheCount++
			}
		},
	})

	if cli.HasSynced() {
		t.Errorf("client not start cache not synced")
		return
	}

	go cli.Start(context.TODO())

	if !cli.HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}
	time.Sleep(time.Second * 2)

	if podCount != cacheCount {
		t.Errorf("%s pod count KubeInterface got %d EventHandler got %d", testNs, podCount, cacheCount)
	}
}

func TestNewClientException(t *testing.T) {
	t.Run("with empty cluster name", func(t *testing.T) {
		_, err := NewClient(
			"",
			WithRuntimeManagerOptions(manager.Options{
				Scheme: scheme,
			}),
		)
		if err == nil {
			t.Error("cluster name is empty")
		}
	})

	t.Run("with error kubeconfig", func(t *testing.T) {
		_, err := NewClient(
			"test-cluster",
			WithKubeConfig(errKubeConfig),
			WithKubeConfigType(KubeConfigTypeRawString),
		)
		if err == nil {
			t.Error("kubeconfig is error")
		}
	})

	t.Run("without health check", func(t *testing.T) {
		cli, err := NewClient(
			"test-cluster",
			WithAutoHealthCheckInterval(0),
		)
		if err != nil {
			t.Error(err)
			return
		}
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*2)
		defer cancel()

		err = cli.Start(ctx)
		if err != nil {
			t.Error(err)
			return
		}
	})
}
