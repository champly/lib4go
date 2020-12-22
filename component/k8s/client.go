package k8s

import (
	"context"
	"errors"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	controllers "sigs.k8s.io/controller-runtime"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Client wrap controller-runtime client
type Client struct {
	*option

	KubeRestConfig *rest.Config
	KubeInterface  kubernetes.Interface

	CtrRtManager manager.Manager
}

// NewClient build Client
func NewClient(opts ...Option) (*Client, error) {
	defaultCfg := getDefaultCfg()
	for _, opt := range opts {
		opt(defaultCfg)
	}
	cli := &Client{option: defaultCfg}

	if err := cli.precheck(); err != nil {
		return nil, err
	}

	if err := cli.initialization(); err != nil {
		return nil, err
	}

	go cli.autoCheck()

	return cli, nil
}

// precheck pre check config
func (cli *Client) precheck() error {
	if cli.clsname == "" {
		return errors.New("cluster name is empty")
	}
	return nil
}

// initialization initialization Client
func (cli *Client) initialization() error {
	var err error
	// Step 1. build restconfig
	cli.KubeRestConfig, err = buildClientCmd(cli.kubeconfig, cli.rsFns)
	if err != nil {
		return err
	}

	// Step 2. build kubernetes interface
	cli.KubeInterface, err = buildKubeInterface(cli.KubeRestConfig)
	if err != nil {
		return err
	}

	// Step 3. build controller-runtime manager
	cli.CtrRtManager, err = controllers.NewManager(cli.KubeRestConfig, cli.rtManagerOpts)
	if err != nil {
		return err
	}

	return nil
}

// autoCheck auto check Client connect status
func (cli *Client) autoCheck() {
	for {
		if cli.Status != Initing {
			time.Sleep(time.Second * 5)
		}

		ok, err := healthRequest(cli.KubeInterface, time.Second*5)
		if err != nil {
			klog.Errorf("cluster [%s] check failed:%+v", cli.clsname, err)
		}
		if !ok {
			cli.Status = DisConnected
			continue
		}

		cli.Status = Connected
	}
}

// Start start client
func (cli *Client) Start(ctx context.Context) error {
	var err error

	ch := make(chan struct{}, 0)
	go func() {
		err = cli.CtrRtManager.Start(ctx)
		close(ch)
	}()

	select {
	case <-cli.stopCh:
		// invoke Stop
		close(cli.gracefulStopCh)
		return nil
	case <-ch:
		// controller-manager stop
		return err
	}
}

// Stop stop client with timeout 30s
func (cli *Client) Stop(ctx context.Context) {
	close(cli.stopCh)

	select {
	case <-cli.gracefulStopCh:
		return
	case <-time.After(GracefulStopWaitTimeout):
		return
	}
}

// AddEventHandler add event handler
func (cli *Client) AddEventHandler(handler cache.ResourceEventHandler, obj client.Object) error {
	informer, err := cli.GetInformerWithObj(obj)
	if err != nil {
		return err
	}
	informer.AddEventHandler(handler)
	return nil
}

// GetInformerWithObj get object informer with cache
func (cli *Client) GetInformerWithObj(obj client.Object) (runtimecache.Informer, error) {
	return cli.CtrRtManager.GetCache().GetInformer(context.TODO(), obj)
}
