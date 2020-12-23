package k8s

import (
	"context"
	"testing"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	networkingv1beta1.AddToScheme(scheme)
}

func TestNewClient(t *testing.T) {
	cli, err := NewClient(
		WithRuntimeManagerOptions(manager.Options{
			Scheme: scheme,
		}),
		WithClusterName("test-cluster"),
	)
	if err != nil {
		t.Error(err)
		return
	}
	defer cli.Stop()

	// informer, err := cli.CtrRtManager.GetCache().GetInformer(context.TODO(), &networkingv1beta1.DestinationRule{})
	// if err != nil {
	//     t.Error(err)
	//     return
	// }
	// informer.AddEventHandler(cache.ResourceEventHandlerFuncs{})
	err = cli.AddEventHandler(&networkingv1beta1.DestinationRule{}, cache.ResourceEventHandlerFuncs{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		if err := cli.Start(context.TODO()); err != nil {
			t.Log(err)
		}
	}()

	// for !informer.HasSynced() {
	//     time.Sleep(time.Second * 1)
	// }

	time.Sleep(time.Second * 2)

	ds := &networkingv1beta1.DestinationRule{}
	if err = cli.CtrRtManager.GetCache().Get(context.TODO(), types.NamespacedName{Namespace: "sym-admin", Name: "com.dmall.bmservice.seq-66"}, ds); err != nil {
		t.Error(err)
		return
	}
	t.Log(ds)
}
