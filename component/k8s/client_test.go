package k8s

import (
	"testing"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Errorf("build client err:%+v", err)
		return
	}

	// obj, err := client.KubeDynamicClient.Resource(networkingv1beta1.SchemeGroupVersion.WithResource("destinationrules")).Namespace(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{Limit: 20})
	// if err != nil {
	//     t.Errorf("result get failed:%+v", err)
	//     return
	// }
	// dsList := &networkingv1beta1.DestinationRuleList{}
	// err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), dsList)
	// if err != nil {
	//     t.Errorf("result get failed:%+v", err)
	//     return
	// }
	// for _, ds := range dsList.Items {
	//     t.Logf("get ds %s/%s", ds.Namespace, ds.Name)
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

	dsInformer := client.DynamicSharedInformerFactory.ForResource(networkingv1beta1.SchemeGroupVersion.WithResource("destinationrules")).Informer()
	dsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// us, ok := obj.(*unstructured.Unstructured)
			// if ok {
			//     ds := &networkingv1beta1.DestinationRule{}
			//     if err := runtime.DefaultUnstructuredConverter.FromUnstructured(us.UnstructuredContent(), ds); err != nil {
			//         t.Errorf("unstructured ds failed:%+v", err)
			//     }
			//     t.Logf("add ds:%s/%s", ds.Namespace, ds.Name)
			// }
		},
		UpdateFunc: func(old, new interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	client.SharedInformerFactory.Start(stopCh)
	client.DynamicSharedInformerFactory.Start(stopCh)

	for !informer.HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}
	t.Log("sync pod success")

	for !dsInformer.HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}
	t.Log("sync destinationrules success")

	ds := &networkingv1beta1.DestinationRule{}
	ds.Name = "com.dmall.bmservice.seq-66"
	ds.Namespace = "sym-admin"
	item, exist, err := dsInformer.GetIndexer().Get(ds)
	t.Logf("item:%+v", item)
	t.Logf("exist:%+v", exist)
	t.Logf("err:%+v", err)
	if err == nil && exist {
		dd := &networkingv1beta1.DestinationRule{}
		us, ok := item.(*unstructured.Unstructured)
		if ok {
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(us.UnstructuredContent(), dd)
			if err == nil {
				t.Logf("ds:%+v", dd)
			}
		}
	}

	time.Sleep(time.Second * 10)
}
