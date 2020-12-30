package k8s

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildClientCmd(kubeConfigType KubeConfigType, kubeconfig string, kubecontext string, setRestConfigFnList []SetKubeRestConfigFn) (*rest.Config, error) {
	switch kubeConfigType {
	case KubeConfigTypeRawString:
		return buildClientCmdWithRawConfig(kubeconfig, kubecontext, setRestConfigFnList)
	case KubeConfigTypeFile:
		return buildClientCmdWithFile(kubeconfig, kubecontext, setRestConfigFnList)
	default:
		return nil, errors.New("just supoort rawstring and file kubeconfig")
	}
}

func buildClientCmdWithRawConfig(kubeconf string, kubecontext string, setRestConfigFnList []SetKubeRestConfigFn) (*rest.Config, error) {
	if kubeconf == "" {
		return nil, errors.New("kubeconfig is empty")
	}
	apiConfig, err := clientcmd.Load([]byte(kubeconf))
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes API config:%+v", err)
	}

	restcfg, err := clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{CurrentContext: kubecontext}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config from API config:%+v", err)
	}

	for _, fn := range setRestConfigFnList {
		fn(restcfg)
	}
	return restcfg, nil
}

func buildClientCmdWithFile(kubeconf string, kubecontext string, setRestConfigFnList []SetKubeRestConfigFn) (*rest.Config, error) {
	if kubeconf != "" {
		info, err := os.Stat(kubeconf)
		if err != nil || info.Size() == 0 {
			kubeconf = ""
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconf
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: kubecontext,
	}

	restcfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	for _, fn := range setRestConfigFnList {
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

func healthRequestWithTimeout(kubeInterface kubernetes.Interface, timeout time.Duration) (bool, error) {
	// Always return false, when the timeout too small, so must more than 100ms
	if timeout < time.Millisecond*100 {
		return false, errors.New("health request timeout must more than 100ms")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	body, err := kubeInterface.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx).Raw()
	if err != nil {
		return false, err
	}
	return strings.EqualFold(string(body), "ok"), nil
}
