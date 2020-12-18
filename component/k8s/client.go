package k8s

import (
	"os"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	*option

	KubeRestConfig        *rest.Config
	KubeClientSet         *kubernetes.Clientset
	SharedInformerFactory informers.SharedInformerFactory
	Status                ConnectStatus
}

func NewClient(opts ...Option) (*Client, error) {
	defaultConfig := buildDefaultOpt()
	for _, opt := range opts {
		opt(defaultConfig)
	}
	client := &Client{
		option: defaultConfig,
	}

	if err := client.BuildClientCmd(); err != nil {
		return nil, err
	}

	client.BuildClientSet()
	client.BuildInformers()

	return client, nil
}

func (c *Client) BuildClientCmd() (err error) {
	kc := c.kubeconfig
	if kc != "" {
		info, err := os.Stat(c.kubeconfig)
		if err != nil || info.Size() == 0 {
			kc = ""
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	loadingRules.ExplicitPath = kc
	configOverrides := &clientcmd.ConfigOverrides{
		ClusterDefaults: clientcmd.ClusterDefaults,
	}
	c.KubeRestConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return err
	}

	for _, f := range c.fns {
		f(c.KubeRestConfig)
	}
	return nil
}

func (c *Client) BuildClientSet() error {
	c.KubeClientSet = kubernetes.NewForConfigOrDie(c.KubeRestConfig)
	return nil
}

func (c *Client) BuildInformers() error {
	c.SharedInformerFactory = informers.NewSharedInformerFactory(c.KubeClientSet, c.resync)
	return nil
}
