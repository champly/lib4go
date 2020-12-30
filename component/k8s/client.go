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

	CtrlRtManager manager.Manager
	CtrlRtClient  client.Client
	CtrlRtCache   runtimecache.Cache
	InformerList  []runtimecache.Informer
}

// NewClient build Client
func NewClient(opts ...Option) (*Client, error) {
	cfg := buildDefaultCfg()
	for _, opt := range opts {
		opt(cfg)
	}
	cli := &Client{option: cfg, InformerList: []runtimecache.Informer{}}

	if err := cli.preCheck(); err != nil {
		return nil, err
	}

	if err := cli.initialization(); err != nil {
		return nil, err
	}

	go cli.autoHealthCheck()

	return cli, nil
}

// preCheck pre check config
func (cli *Client) preCheck() error {
	// cluster name must not empty
	if cli.GetName() == "" {
		return errors.New("cluster name is empty")
	}
	return nil
}

// initialization initialization Client
func (cli *Client) initialization() error {
	var err error
	// Step 1. build restconfig
	cli.KubeRestConfig, err = buildClientCmd(cli.kubeConfigType, cli.kubeConfig, cli.kubeContext, cli.setKubeRestConfigFnList)
	if err != nil {
		return fmt.Errorf("cluster [%s] build kubernetes restConfig with type %s failed:%+v", cli.GetName(), cli.kubeConfigType, err)
	}

	// Step 2. build kubernetes interface
	cli.KubeInterface, err = buildKubeInterface(cli.KubeRestConfig)
	if err != nil {
		return fmt.Errorf("cluster [%s] build kubernetes interface failed:%+v", cli.GetName(), err)
	}

	// Step 3. build controller-runtime manager
	cli.CtrlRtManager, err = controllers.NewManager(cli.KubeRestConfig, cli.ctrlRtManagerOpts)
	if err != nil {
		return fmt.Errorf("cluster [%s] build controller-runtime manager failed:%+v", cli.GetName(), err)
	}
	cli.CtrlRtClient = cli.CtrlRtManager.GetClient()
	cli.CtrlRtCache = cli.CtrlRtManager.GetCache()

	return nil
}

// autoHealthCheck auto check Client connect status
func (cli *Client) autoHealthCheck() {
	if cli.healthCheckInterval <= 0 {
		return
	}

	for {
		if cli.ConnectStatus != Initing {
			time.Sleep(cli.healthCheckInterval)
		}

		ok, err := healthRequestWithTimeout(cli.KubeInterface, time.Second*5)
		if err != nil {
			// just log error
			klog.Errorf("cluster [%s] check healthy failed:%+v", cli.GetName(), err)
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
		return fmt.Errorf("client %s can't repeat start", cli.GetName())
	}
	cli.StartStatus = true

	ctx, cancel := context.WithCancel(ctx)
	cli.cancel = cancel

	var err error
	ch := make(chan struct{}, 0)
	go func() {
		err = cli.CtrlRtManager.Start(ctx)
		if err != nil {
			klog.Errorf("start cluster [%s] error:%+v", cli.GetName(), err)
		}
		close(ch)
	}()

	select {
	case <-ch:
		// controller-manager stop
		close(cli.stopCh)
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
	case <-cli.stopCh:
		klog.Infof("cluster %s has been stopped", cli.GetName())
		return
	case <-time.After(GracefulStopWaitTimeout):
		klog.Errorf("stop cluster %s timeout", cli.GetName())
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
	informer, err := cli.CtrlRtCache.GetInformer(context.TODO(), obj)
	if err != nil {
		return nil, fmt.Errorf("cluster %s GetInformerWithObj error:%+v", cli.GetName(), err)
	}
	cli.InformerList = append(cli.InformerList, informer)
	return informer, nil
}

// HasSynced return all informer has synced
func (cli *Client) HasSynced() bool {
	if !cli.StartStatus {
		// if not start, the informer will not synced
		return false
	}

	for _, informer := range cli.InformerList {
		if !informer.HasSynced() {
			return false
		}
	}
	return true
}

// GetName return cluster name
func (cli *Client) GetName() string {
	return cli.clsName
}
