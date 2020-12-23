package k8s

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	kubecfgFileBakSuffix = ".bak"
	autoRebuildInterval  = time.Second * 2
	waitRebuildtime      = time.Second * 3
	filterNamespace      = "sym-admin"
)

func TestMulitCluster(t *testing.T) {
	tempYamlToJson()

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

	mulitClient, err := NewMulitClient(autoRebuildInterval, cdcfg)
	if err != nil {
		t.Errorf("build mulit client failed:%+v", err)
		return
	}

	mulitClient.BeforeStartFunc = func(cli *Client) error {
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
	}

	mulitClient.Start(context.TODO())
	defer mulitClient.Stop()

	for i := 0; i < 10; i++ {
		t.Log("add one by one")
		for addOneKube() {
			time.Sleep(waitRebuildtime)
			if !mulitClient.HasSynced() {
				time.Sleep(time.Microsecond * 100)
			}
		}

		t.Log("remove one by one")
		for removeOnKube() {
			time.Sleep(waitRebuildtime)
			if !mulitClient.HasSynced() {
				time.Sleep(time.Microsecond * 100)
			}
		}
	}
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
