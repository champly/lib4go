package k8s

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildClientCmd(kubecfg string, context string, rsFns []RestConfigFunc) (*rest.Config, error) {
	if kubecfg != "" {
		info, err := os.Stat(kubecfg)
		if err != nil || info.Size() == 0 {
			kubecfg = ""
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	loadingRules.ExplicitPath = kubecfg
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
		CurrentContext:  context,
	}

	var (
		err     error
		restcfg *rest.Config
	)
	restcfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	for _, fn := range rsFns {
		fn(restcfg)
	}
	return restcfg, nil
}

func buildKubeInterface(restcfg *rest.Config) (kubernetes.Interface, error) {
	clientSet, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

func healthRequest(kubeInterface kubernetes.Interface, timeout time.Duration) (bool, error) {
	// Always return false, when the timeout too small, so must more than 100ms
	if timeout < time.Millisecond*100 {
		return false, errors.New("timeout must more than 100ms")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	body, err := kubeInterface.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx).Raw()
	if err != nil {
		return false, err
	}
	return strings.EqualFold(string(body), "ok"), nil
}
