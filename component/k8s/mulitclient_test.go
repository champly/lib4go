package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestMulitCluster(t *testing.T) {
	cli, err := NewClient(
		WithClusterName("test-cluster-configmap"),
		WithRuntimeManagerOptions(manager.Options{
			MetricsBindAddress: "0",
		}),
	)
	if err != nil {
		t.Errorf("build client failed:%+v", err)
		return
	}
	defer cli.Stop()

	cdcfg, err := NewClusterCfgWithDir(
		cli.KubeInterface,
		clsConfigurationTmpDir,
		clsConfigurationSuffix,
		WithRuntimeManagerOptions(
			manager.Options{
				MetricsBindAddress:     "0",
				LeaderElection:         false,
				HealthProbeBindAddress: "0",
			},
		),
		WithResetConfigFunc(
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

	mulitClient, err := NewMulitClient(cdcfg)
	if err != nil {
		t.Errorf("build mulit client failed:%+v", err)
		return
	}

	mulitClient.BeforeStartFunc = func(cli *Client) error {
		cli.GetInformerWithObj(&corev1.Pod{})
		return nil
	}

	go func() {
		mulitClient.Start(context.TODO())
	}()
	defer mulitClient.Stop()

	for !mulitClient.HasSynced() {
		time.Sleep(time.Millisecond * 100)
	}

	for _, cli := range mulitClient.GetAll() {
		list := &corev1.PodList{}
		err = cli.CtrRtManager.GetCache().List(context.TODO(), list, &client.ListOptions{Namespace: "sym-admin"})
		if err != nil {
			t.Errorf("get cluster %s podlist failed:%+v", cli.GetName(), err)
			return
		}

		for _, pod := range list.Items {
			t.Logf("%s %s/%s", cli.GetName(), pod.Namespace, pod.Name)
		}
	}
}
