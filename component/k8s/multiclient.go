package k8s

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MultiClient multi cluster client obj
type MultiClient struct {
	ctx             context.Context
	started         bool
	stopCh          chan struct{}
	clusterCfgMgr   ClusterConfigurationManager
	clusterCliMap   map[string]*Client
	rebuildInterval time.Duration
	l               sync.Mutex
	InitHandlerList []InitHandler
}

// NewMultiClient build MultiClient
func NewMultiClient(rebuildInterval time.Duration, clusterCfgMgr ClusterConfigurationManager) (*MultiClient, error) {
	multiCli := &MultiClient{
		rebuildInterval: rebuildInterval,
		stopCh:          make(chan struct{}),
		clusterCfgMgr:   clusterCfgMgr,
		clusterCliMap:   map[string]*Client{},
		InitHandlerList: []InitHandler{},
	}

	clsList, err := multiCli.clusterCfgMgr.GetAll()
	if err != nil {
		return nil, fmt.Errorf("get all cluster info failed:%+v", err)
	}

	for _, clsInfo := range clsList {
		cli, err := buildClient(clsInfo, multiCli.clusterCfgMgr.GetOptions()...)
		if err != nil {
			return nil, err
		}
		multiCli.clusterCliMap[clsInfo.GetName()] = cli
	}

	return multiCli, nil
}

func buildClient(clsInfo ClusterConfigInfo, options ...Option) (*Client, error) {
	opts := []Option{}
	opts = append(opts, WithKubeConfig(clsInfo.GetKubeConfig()))
	opts = append(opts, WithKubeContext(clsInfo.GetKubeContext()))
	opts = append(opts, WithKubeConfigType(clsInfo.GetKubeConfigType()))
	opts = append(opts, options...)

	return NewClient(clsInfo.GetName(), opts...)
}

// AddEventHandler add event with multiclient
func (mc *MultiClient) AddEventHandler(obj client.Object, handler cache.ResourceEventHandler) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for name, cli := range mc.clusterCliMap {
		err = cli.AddEventHandler(obj, handler)
		if err != nil {
			return fmt.Errorf("cluster [%s] AddEventHandler failed:%+v", name, err)
		}
	}
	return nil
}

// TriggerSync only trigger informer sync obj
func (mc *MultiClient) TriggerSync(obj client.Object) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for name, cli := range mc.clusterCliMap {
		_, err = cli.GetInformerWithObj(obj)
		if err != nil {
			return fmt.Errorf("cluster [%s] TriggerObjSync failed:%+v", name, err)
		}
	}
	return nil
}

// SetIndexField set informer indexfield
func (mc *MultiClient) SetIndexField(obj client.Object, field string, extractValue client.IndexerFunc) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for name, cli := range mc.clusterCliMap {
		err = cli.CtrlRtManager.GetFieldIndexer().IndexField(context.TODO(), obj, field, extractValue)
		if err != nil {
			return fmt.Errorf("cluster [%s] SetIndexField [%s] failed:%+v", name, field, err)
		}
	}
	return nil
}

// GetWithName get cluster with name.
func (mc *MultiClient) GetWithName(name string) (*Client, error) {
	mc.l.Lock()
	defer mc.l.Unlock()

	cli, ok := mc.clusterCliMap[name]
	if !ok {
		return nil, fmt.Errorf("cluster [%s] not found, maybe not registry", name)
	}
	return cli, nil
}

// GetConnectedWithName get cluster with name and cluster is healthy.
func (mc *MultiClient) GetConnectedWithName(name string) (*Client, error) {
	cli, err := mc.GetWithName(name)
	if err != nil {
		return nil, err
	}
	if cli.ConnectStatus != Connected {
		return nil, fmt.Errorf("cluster [%s] not connected apiserver", name)
	}
	return cli, nil
}

// GetReadyWithName get cluster with name and cluster is healthy and status ready.
func (mc *MultiClient) GetReadyWithName(name string) (*Client, error) {
	cli, err := mc.GetWithName(name)
	if err != nil {
		return nil, err
	}
	if cli.ConnectStatus == Connected && cli.HasSynced() {
		return cli, nil
	}
	return nil, fmt.Errorf("cluster [%s] not connected apiserver or not ready", name)
}

// GetAllConnected get all cluster when cluster is connected.
func (mc *MultiClient) GetAllConnected() []*Client {
	mc.l.Lock()
	defer mc.l.Unlock()

	cliList := make([]*Client, 0, len(mc.clusterCliMap))
	for _, cli := range mc.clusterCliMap {
		if cli.ConnectStatus == Connected {
			cliList = append(cliList, cli)
		}
	}
	return cliList
}

// GetAllReady get all cluster when cluster is connected and informer has synced
func (mc *MultiClient) GetAllReady() []*Client {
	mc.l.Lock()
	defer mc.l.Unlock()

	cliList := make([]*Client, 0, len(mc.clusterCliMap))
	for _, cli := range mc.clusterCliMap {
		if cli.ConnectStatus == Connected && cli.HasSynced() {
			cliList = append(cliList, cli)
		}
	}
	return cliList
}

// GetAll get all cluster.
func (mc *MultiClient) GetAll() []*Client {
	mc.l.Lock()
	defer mc.l.Unlock()

	cliList := make([]*Client, 0, len(mc.clusterCliMap))
	for _, cli := range mc.clusterCliMap {
		cliList = append(cliList, cli)
	}
	return cliList
}

// HasSynced return all cluster has synced
func (mc *MultiClient) HasSynced() bool {
	mc.l.Lock()
	defer mc.l.Unlock()

	for _, cli := range mc.clusterCliMap {
		if !cli.HasSynced() {
			return false
		}
	}
	return true
}

// Start start multiclient
func (mc *MultiClient) Start(ctx context.Context) error {
	if mc.started {
		return errors.New("not restart multiclient")
	}
	mc.l.Lock()
	defer mc.l.Unlock()

	mc.started = true
	go mc.autoRebuild()

	mc.ctx = ctx
	var err error
	for _, cli := range mc.clusterCliMap {
		err = startClient(mc.ctx, cli, mc.InitHandlerList)
		if err != nil {
			return err
		}
	}
	return nil
}

func startClient(ctx context.Context, cli *Client, initHandlerList []InitHandler) error {
	for _, handler := range initHandlerList {
		err := handler(cli)
		if err != nil {
			return fmt.Errorf("invoke cluster [%s] InitHandler failed:%+v", cli.GetName(), err)
		}
	}

	go cli.Start(ctx)

	return nil
}

// Stop multiclient
func (mc *MultiClient) Stop() {
	if !mc.started {
		return
	}

	close(mc.stopCh)

	wg := &sync.WaitGroup{}
	for _, cli := range mc.clusterCliMap {
		wg.Add(1)
		go func(cli *Client) {
			cli.Stop()
			wg.Done()
		}(cli)
	}
	wg.Wait()
}

func (mc *MultiClient) autoRebuild() {
	if mc.rebuildInterval <= 0 {
		return
	}

	var err error
	for {
		select {
		case <-mc.stopCh:
			return
		case <-time.After(mc.rebuildInterval):
			err = mc.Rebuild()
			if err != nil {
				klog.Errorf("Rebuild failed:%+v", err)
			}
		}
	}
}

// Rebuild rebuild with cluster info
func (mc *MultiClient) Rebuild() error {
	if !mc.started {
		return nil
	}

	mc.l.Lock()
	defer mc.l.Unlock()

	newClsList, err := mc.clusterCfgMgr.GetAll()
	if err != nil {
		return fmt.Errorf("get all cluster info failed:%+v", err)
	}

	newCliMap := make(map[string]*Client, len(newClsList))
	// add and check new cluster
	for _, newcls := range newClsList {
		// get old client info
		oldcli, exist := mc.clusterCliMap[newcls.GetName()]
		if exist && oldcli.kubeConfig == newcls.GetKubeConfig() {
			// if kubeconfig not modify
			newCliMap[oldcli.GetName()] = oldcli
			continue
		}

		// build new client
		cli, err := buildClient(newcls, mc.clusterCfgMgr.GetOptions()...)
		if err != nil {
			klog.Error(err)
			continue
		}

		// start new client
		err = startClient(mc.ctx, cli, mc.InitHandlerList)
		if err != nil {
			return err
		}

		if exist {
			// kubeconfig modify, should stop old client
			oldcli.Stop()
		}

		newCliMap[cli.GetName()] = cli
		klog.Infof("auto add cluster %s", cli.GetName())
	}

	// remove unexpect cluster
	for name, oldcli := range mc.clusterCliMap {
		if _, ok := newCliMap[name]; !ok {
			// not exist, should stop
			go func(cli *Client) {
				cli.Stop()
			}(oldcli)
		}
	}

	mc.clusterCliMap = newCliMap
	return nil
}
