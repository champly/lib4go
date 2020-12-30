package k8s

import (
	"os"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	clsConfigurationNamespace = "sym-admin"
	clsConfigurationLabel     = map[string]string{
		"ClusterOwner": "sym-admin",
	}
	clsConfigurationDataname = "kubeconfig.yaml"

	clsConfigurationTmpDir = "/tmp/kubedir"
	clsConfigurationSuffix = "yaml"
)

func TestClusterCfgWithCM(t *testing.T) {
	cli, err := NewClient(
		"test-cluster-configmap",
		WithRuntimeManagerOptions(manager.Options{
			MetricsBindAddress: "0",
		}),
	)
	if err != nil {
		t.Errorf("build client failed:%+v", err)
		return
	}
	defer cli.Stop()

	cc := NewClusterCfgWithCM(cli.KubeInterface, clsConfigurationNamespace, clsConfigurationLabel, clsConfigurationDataname)
	clsList, err := cc.GetAll()
	if err != nil {
		t.Error(err)
		return
	}
	for _, cls := range clsList {
		t.Logf("%s found", cls.GetName())
	}
	return
}

func TestClusterCfgWithDir(t *testing.T) {
	err := buildTmpWithKubeConfig()
	if err != nil {
		t.Errorf("buildtempwithkubeconfig failed:%+v", err)
		return
	}

	cli, err := NewClient(
		"test-cluster-configmap",
		WithRuntimeManagerOptions(manager.Options{
			MetricsBindAddress: "0",
		}),
	)
	if err != nil {
		t.Errorf("build client failed:%+v", err)
		return
	}
	defer cli.Stop()

	cd, err := NewClusterCfgWithDir(cli.KubeInterface, clsConfigurationTmpDir, clsConfigurationSuffix, KubeConfigTypeFile)
	if err != nil {
		t.Errorf("build cluster configuration with dir failed:%+v", err)
		return
	}
	clsList, err := cd.GetAll()
	if err != nil {
		t.Error(err)
		return
	}
	for _, cls := range clsList {
		t.Logf("%s found", cls.GetName())
	}
	return
}

func buildTmpWithKubeConfig() error {
	home, _ := os.UserHomeDir()
	path := home + "/.kube/config"
	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return err
	}
	for name, context := range config.Contexts {
		if !strings.Contains(name, "test-bus") {
			continue
		}
		cfg := clientcmdapi.NewConfig()
		cfg.APIVersion = config.APIVersion
		cfg.CurrentContext = name
		cfg.Clusters[context.Cluster] = config.Clusters[context.Cluster]
		cfg.AuthInfos[context.AuthInfo] = config.AuthInfos[context.AuthInfo]
		cfg.Contexts[name] = config.Contexts[name]

		clientcmd.WriteToFile(*cfg, clsConfigurationTmpDir+"/"+name+"."+clsConfigurationSuffix)
	}
	return nil
}
