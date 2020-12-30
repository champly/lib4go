package k8s

import (
	"context"
	"testing"
	"time"

	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme = runtime.NewScheme()
)

var (
	errKubeConfig = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01ESXhNREF6TlRNd01sb1hEVE13TURJd056QXpOVE13TWxvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTXVZCkhraWJHRnhtREpYN3RqdDhhL1BxWU56aTRadlRpZU9HdjRLWkJrL09JMUxPL1NacDBaaUZrNE1md1hxY2RST3YKM2dTblBhSVZQMGVNQUlCMDRGdUtHck9aRFk3djByQU5GeHBqZk8xVDVCeXRGZkJuUzNIeUNmMTgvZlVyU3NCWQpNc1gxY2p1N0VhZFVJZDdnSktiQndhdUxLYnhWWmp1K2ZQQys2VTVyQVpITGt2eVNOeEd3a1BDOGxNMFA0Um1BCmdvem5QbnJVVUVtVTFWN1gvUDZjZzg0WTBPNFhuaENsZS9ySmpEVXQ4dHNKTFg3dkdLL3diSUhSclBrUUNIN2YKWDBGMEFtV0htaHI5c0RZYnVoRlU4bHA4eGhkb24reVY5U3JXME1CWGFLcUtpN2U2VnBjWXR6ZWMzZEhwSmh4SQpJb0U3Wi8zU1JrcGh3bFZOblBrQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0tVTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFGNTBiR3g2ZTdDMGVJVDYyT0FRN3J1d1pSblEKcnFnWkVGNUtzd3FTSDZqRmJtN1dGMFpqb2s1b21uMVBaVHUwSzhSOHN3cDFLZ3lTMGJKamIxWWJ5WUF6eERGWApSdVRKbGIvYWlucGRLTFhRQWZCaSs4ZTdmTDlPamd2aDlDMjFUSE5YUVpBTWNDZThSdVZubERYUTBHRGgyYWJnCnRQZlU2MlV5cUp5MnMyUWlkM3BocmNkRmNuWUNHUG5WTy8yWmkxNGJlYTdOdEdDc2N5ZDRzTURqUlhlU0oyZGQKdHhZWWNQeFpFbWVYOUF0V00wU045ZVRLZ2l0TmZxV202RHFTakN1NGN3eEk3WW5rWjlOQ3h1OHhaZ2FlZTJUYgptSFZGOTgyT3BJY3d6MVcvM0ZpL1lJcXRQb2R5dmtWZGRQS29QdDJjOGJFS3BtWFRHZUpobkhXWTMxUT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://cls-otdyiqyb.ccs.tencent-cloud.com
  name: test-tke-gz-bj5-bus-01
contexts: null
current-context: test-bus-gz-01-tke-bj5
kind: Config
preferences: {}
users:
- name: test-tke-gz-bj5-bus-01-admin
  user:
    token: iL0fOuAoevn4fUivczhfY2ZHduOVMbEL
	`
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = networkingv1beta1.AddToScheme(scheme)
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

	cli.AddEventHandler(&corev1.Pod{}, cache.ResourceEventHandlerFuncs{})
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		if err := cli.Start(context.TODO()); err != nil {
			t.Log(err)
		}
	}()

	for !cli.HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}
}

func TestNewClientException(t *testing.T) {
	t.Run("without name", func(t *testing.T) {
		_, err := NewClient(
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
			WithKubeConfig(errKubeConfig),
			WithClusterName("test-cluster"),
			WithKubeConfigType(KubeConfigTypeRawString),
		)
		if err == nil {
			t.Error("kubeconfig is error")
		}
	})

	t.Run("without health check", func(t *testing.T) {
		cli, err := NewClient(
			WithAutoHealthCheckInterval(0),
			WithClusterName("test-cluster"),
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
