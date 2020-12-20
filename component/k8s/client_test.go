package k8s

import (
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Errorf("build client err:%+v", err)
		return
	}

	// pods, err := client.KubeClientSet.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})
	// if err != nil {
	//     t.Errorf("get all pods err:%+v", err)
	//     return
	// }
	// for _, pod := range pods.Items {
	//     t.Logf("get pod %s/%s ", pod.Namespace, pod.Name)
	// }

	informer := client.SharedInformerFactory.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// t.Logf("add obj:%+v", obj)
		},
		UpdateFunc: func(old, new interface{}) {
			// t.Logf("update obj: %+v, %+v", old, new)
		},
		DeleteFunc: func(obj interface{}) {
			// t.Logf("delete obj:%+v", obj)
		},
	})

	// client.SharedInformerFactory.InformerFor(nil, nil)

	// genericInformer, err := client.SharedInformerFactory.ForResource(schema.GroupVersionResource{
	//     Group:    "group",
	//     Version:  "version",
	//     Resource: "resource",
	// })
	// if err != nil {
	//     t.Errorf("build informer for resource failed:%+v", err)
	//     return
	// }

	// genericInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//     AddFunc: func(obj interface{}) {
	//         // t.Logf("add obj:%+v", obj)
	//     },
	//     UpdateFunc: func(old, new interface{}) {
	//         // t.Logf("update obj: %+v, %+v", old, new)
	//     },
	//     DeleteFunc: func(obj interface{}) {
	//         // t.Logf("delete obj:%+v", obj)
	//     },
	// })
	client.SharedInformerFactory.InformerFor(nil, NewInformerFunc)

	stopCh := make(chan struct{})
	client.SharedInformerFactory.Start(stopCh)

	for !informer.HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}
	t.Log("sync pod success")

	// for !genericInformer.Informer().HasSynced() {
	//     time.Sleep(time.Millisecond * 100)
	// }
	// t.Log("sync appsets success")

	time.Sleep(time.Second * 10)
}

func NewInformerFunc(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			// ListFunc: func(options metav1.ListOptions)(runtime.Object, error){
			//     return client.AppsV1().Deployments().Watch
			// },
		},
		nil,
		resyncPeriod,
		nil,
	)
}
