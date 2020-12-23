package k8s

import (
	"context"
	"errors"
	"fmt"
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

	cancel context.CancelFunc

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
	// cluster name must not empty
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
		return fmt.Errorf("build kubernetes restconfig failed:%+v", err)
	}

	// Step 2. build kubernetes interface
	cli.KubeInterface, err = buildKubeInterface(cli.KubeRestConfig)
	if err != nil {
		return fmt.Errorf("build kubernetes interface failed:%+v", err)
	}

	// Step 3. build controller-runtime manager
	cli.CtrRtManager, err = controllers.NewManager(cli.KubeRestConfig, cli.rtManagerOpts)
	if err != nil {
		return fmt.Errorf("build controller-runtime manager failed:%+v", err)
	}

	return nil
}

// autoCheck auto check Client connect status
func (cli *Client) autoCheck() {
	for {
		if cli.ConnectStatus != Initing {
			time.Sleep(time.Second * 5)
		}

		ok, err := healthRequest(cli.KubeInterface, time.Second*5)
		if err != nil {
			klog.Errorf("cluster [%s] check failed:%+v", cli.clsname, err)
		}
		if !ok {
			cli.ConnectStatus = DisConnected
			continue
		}

		cli.ConnectStatus = Connected
	}
}

// Start start client
func (cli *Client) Start(ctx context.Context) error {
	if cli.StartStatus {
		return fmt.Errorf("client %s can't repeat start", cli.clsname)
	}

	ctx, cancel := context.WithCancel(ctx)
	cli.cancel = cancel

	var err error
	ch := make(chan struct{}, 0)
	go func() {
		cli.StartStatus = true
		err = cli.CtrRtManager.Start(ctx)
		close(ch)
	}()

	select {
	case <-ch:
		// controller-manager stop
		close(cli.gracefulStopCh)
		return err
	}
}

// Stop stop client with timeout 30s
func (cli *Client) Stop() {
	if !cli.StartStatus || cli.cancel == nil {
		return
	}

	cli.cancel()

	select {
	case <-cli.gracefulStopCh:
		return
	case <-time.After(GracefulStopWaitTimeout):
		klog.Errorf("close cluster %s timeout", cli.clsname)
		return
	}
}

// AddEventHandler add event handler
func (cli *Client) AddEventHandler(obj client.Object, handler cache.ResourceEventHandler) error {
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
