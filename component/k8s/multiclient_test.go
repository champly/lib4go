package k8s

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	kubecfgFileBakSuffix = ".bak"
	autoRebuildInterval  = time.Second * 2
	waitRebuildtime      = time.Second * 3
	filterNamespace      = "default"
)

func TestMultiClusterWithCM(t *testing.T) {
	buildTmpWithKubeConfig()
	defer clean()

	cli, err := NewClient("test-cluster-configmap")
	if err != nil {
		t.Errorf("build client failed:%+v", err)
		return
	}
	defer cli.Stop()

	clusterMgr, err := NewClusterCfgManagerWithDir(
		clsConfigurationTmpDir,
		clsConfigurationSuffix,
		KubeConfigTypeFile,
		WithRuntimeManagerOptions(
			manager.Options{
				MetricsBindAddress:     "0",
				LeaderElection:         false,
				HealthProbeBindAddress: "0",
			},
		),
		WithKubeSetRsetConfigFn(
			func(restcfg *rest.Config) {
				restcfg.QPS = 100
				restcfg.Burst = 120
			},
		),
	)
	if err != nil {
		t.Errorf("build cluster configuration with dir failed:%+v", err)
		return
	}

	multiClient, err := NewMultiClient(autoRebuildInterval, clusterMgr)
	if err != nil {
		t.Errorf("build multi client failed:%+v", err)
		return
	}

	multiClient.InitHandlerList = append(multiClient.InitHandlerList, func(cli *Client) error {
		cli.AddEventHandler(&corev1.Pod{}, cache.ResourceEventHandlerFuncs{
			// AddFunc    func(obj interface{})
			// UpdateFunc func(oldObj, newObj interface{})
			// DeleteFunc func(obj interface{})
			AddFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err != nil {
					t.Errorf("transfor failed:%+v", err)
					return
				}
				ns, name, err := cache.SplitMetaNamespaceKey(key)
				if err != nil {
					t.Errorf("split namespace failed:%+v", err)
					return
				}
				if ns == filterNamespace {
					t.Logf("%s add %s/%s", cli.GetName(), ns, name)
				}
			},
		})
		return nil
	})

	multiClient.Start(context.TODO())
	defer multiClient.Stop()

	for i := 0; i < 1; i++ {
		t.Log("add one by one")
		for addOneKube() {
			time.Sleep(waitRebuildtime)
			if !multiClient.HasSynced() {
				time.Sleep(time.Microsecond * 100)
			}
		}

		t.Log("remove one by one")
		for removeOnKube() {
			time.Sleep(waitRebuildtime)
			if !multiClient.HasSynced() {
				time.Sleep(time.Microsecond * 100)
			}
		}
	}
}

func TestMultiCluster(t *testing.T) {
	buildTmpWithKubeConfig()
	defer clean()

	cli, err := NewClient("test-cluster-configmap")
	if err != nil {
		t.Errorf("build client failed:%+v", err)
		return
	}
	defer cli.Stop()

	clusterMgr, err := NewClusterCfgManagerWithDir(
		clsConfigurationTmpDir,
		clsConfigurationSuffix,
		KubeConfigTypeFile,
		WithRuntimeManagerOptions(
			manager.Options{
				MetricsBindAddress:     "0",
				LeaderElection:         false,
				HealthProbeBindAddress: "0",
				Scheme:                 scheme,
			},
		),
		WithKubeSetRsetConfigFn(
			func(restcfg *rest.Config) {
				restcfg.QPS = 100
				restcfg.Burst = 120
			},
		),
	)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("without autoRebuild", func(t *testing.T) {
		multi, err := NewMultiClient(0, clusterMgr)
		if err != nil {
			t.Error(err)
			return
		}
		multi.Stop()
	})

	multi, err := NewMultiClient(autoRebuildInterval, clusterMgr)
	if err != nil {
		t.Error(err)
		return
	}
	defer multi.Stop()

	t.Run("AddEventHandler", func(t *testing.T) {
		err := multi.AddEventHandler(&networkingv1beta1.DestinationRule{}, cache.ResourceEventHandlerFuncs{})
		if err == nil {
			t.Error("unregistry type must be error")
		}

		err = multi.AddEventHandler(&corev1.Pod{}, cache.ResourceEventHandlerFuncs{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TriggerObjSync", func(t *testing.T) {
		err := multi.TriggerSync(&networkingv1beta1.DestinationRule{})
		if err == nil {
			t.Error("unregistry type must be error")
		}

		err = multi.TriggerSync(&corev1.Pod{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("SetIndexField", func(t *testing.T) {
		err := multi.SetIndexField(&networkingv1beta1.DestinationRule{}, "name", func(obj client.Object) []string {
			ds := obj.(*networkingv1beta1.DestinationRule)
			return []string{ds.Name}
		})
		if err == nil {
			t.Error("unregistry type must be error")
		}

		err = multi.SetIndexField(&corev1.Pod{}, "status.phase", func(obj client.Object) []string {
			pod := obj.(*corev1.Pod)
			return []string{string(pod.Status.Phase)}
		})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("GetConnectedWithName", func(t *testing.T) {
		_, err := multi.GetConnectedWithName("notfound")
		if err == nil {
			t.Error("notfound cluster must be error")
		}

		clsList := multi.GetAllConnected()
		if len(clsList) < 1 {
			t.Error("connected cluster empty")
			return
		}

		_, err = multi.GetConnectedWithName(clsList[0].GetName())
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		clsList := multi.GetAll()
		if len(clsList) < 1 {
			t.Error("cluster empty")
		}
	})

	t.Run("before start Rebuild", func(t *testing.T) {
		err := multi.Rebuild()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("before start stop", func(t *testing.T) {
		multi.Stop()
	})

	go multi.Start(context.TODO())

	for !multi.HasSynced() {
		t.Log("multiclient wait sync")
		time.Sleep(time.Millisecond * 100)
	}

	t.Run("GetReadyClient", func(t *testing.T) {
		_, err := multi.GetReadyWithName("notfound")
		if err == nil {
			t.Error("notfound cluster must be error")
		}

		clsList := multi.GetAllReady()
		if len(clsList) < 1 {
			t.Error("ready cluster empty")
			return
		}

		_, err = multi.GetReadyWithName(clsList[0].GetName())
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Rebuild", func(t *testing.T) {
		err := multi.Rebuild()
		if err != nil {
			t.Error(err)
		}
	})
}

func tempYamlToJson() {
	files, _ := ioutil.ReadDir(clsConfigurationTmpDir)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), clsConfigurationSuffix) {
			os.Rename(clsConfigurationTmpDir+"/"+file.Name(), clsConfigurationTmpDir+"/"+file.Name()+kubecfgFileBakSuffix)
		}
	}
}

func addOneKube() bool {
	files, _ := ioutil.ReadDir(clsConfigurationTmpDir)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), clsConfigurationSuffix) {
			name := strings.Trim(file.Name(), kubecfgFileBakSuffix)
			os.Rename(clsConfigurationTmpDir+"/"+file.Name(), clsConfigurationTmpDir+"/"+name)
			return true
		}
	}
	return false
}

func removeOnKube() bool {
	files, _ := ioutil.ReadDir(clsConfigurationTmpDir)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), clsConfigurationSuffix) {
			os.Rename(clsConfigurationTmpDir+"/"+file.Name(), clsConfigurationTmpDir+"/"+file.Name()+kubecfgFileBakSuffix)
			return true
		}
	}
	return false
}

func clean() {
	files, _ := ioutil.ReadDir(clsConfigurationTmpDir)
	for _, file := range files {
		os.Remove(file.Name())
	}
}
