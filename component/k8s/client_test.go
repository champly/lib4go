package k8s

import (
	"context"
	"testing"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	networkingv1beta1.AddToScheme(scheme)
}

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Errorf("build client err:%+v", err)
		return
	}

	obj, err := client.KubeDynamicClient.Resource(networkingv1beta1.SchemeGroupVersion.WithResource("destinationrules")).Namespace(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{Limit: 20})
	if err != nil {
		t.Errorf("result get failed:%+v", err)
		return
	}
	dsList := &networkingv1beta1.DestinationRuleList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), dsList)
	if err != nil {
		t.Errorf("result get failed:%+v", err)
		return
	}
	for _, ds := range dsList.Items {
		t.Logf("get ds %s/%s", ds.Namespace, ds.Name)
	}

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
			// runtime.UnstructuredConverter.FromUnstructured()
			us, ok := obj.(*unstructured.Unstructured)
			if ok {
				ds := &networkingv1beta1.DestinationRule{}
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(us.UnstructuredContent(), ds); err != nil {
					t.Errorf("unstructured ds failed:%+v", err)
				}
				t.Logf("add ds:%s/%s", ds.Namespace, ds.Name)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			ds, ok := new.(*networkingv1beta1.DestinationRule)
			if ok {
				t.Logf("update ds:%s/%s", ds.Namespace, ds.Name)
			}
		},
		DeleteFunc: func(obj interface{}) {
			ds, ok := obj.(*networkingv1beta1.DestinationRule)
			if ok {
				t.Logf("delete ds:%s/%s", ds.Namespace, ds.Name)
			}
		},
	})

	stopCh := make(chan struct{})
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

	time.Sleep(time.Second * 10)
}
