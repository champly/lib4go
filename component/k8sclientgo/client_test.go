package k8scligo

import (
	"testing"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Errorf("build client err:%+v", err)
		return
	}

	externalversionsCli, err := clientset.NewForConfig(client.KubeRestConfig)
	if err != nil {
		t.Errorf("build externalversions client err: %+v", err)
		return
	}
	externalSharedInformerFactory := externalversions.NewSharedInformerFactory(externalversionsCli, 0)
	crdInformer := externalSharedInformerFactory.Apiextensions().V1().CustomResourceDefinitions()
	crdInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// us, ok := obj.(*unstructured.Unstructured)
			// if ok {
			//     crd := &apiextensions.CustomResourceDefinition{}
			//     if err := runtime.DefaultUnstructuredConverter.FromUnstructured(us.UnstructuredContent(), crd); err != nil {
			//         t.Errorf("unstructured ds failed:%+v", err)
			//     }
			//     t.Logf("add crd:%s/%s", crd.Namespace, crd.Name)
			// }
		},
	})
	// metaclient, err := metadata.NewForConfig(client.KubeRestConfig)
	// if err != nil {
	//     t.Errorf("build meta client err:%+v", err)
	//     return
	// }
	// metadatainformer.NewSharedInformerFactory(metaclient, 0)

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
	externalSharedInformerFactory.Start(stopCh)

	for !crdInformer.Informer().HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}
	list, err := crdInformer.Lister().List(labels.Everything())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("get crd count: ", len(list))
	for _, item := range list {
		t.Logf("get crd %s -----> %s", item.Spec.Group, item.Name)
	}

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
