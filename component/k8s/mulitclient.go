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

// MulitClient mulit cluster client obj
type MulitClient struct {
	ctx             context.Context
	stopCh          chan struct{}
	BeforeStartFunc BeforeStartFunc
	ClusterCfg      IClusterConfiguration
	ClusterCliMap   map[string]*Client
	reBuildInterval time.Duration
	l               sync.Mutex
	started         bool
}

// NewMulitClient build MulitClient
func NewMulitClient(autoRbTime time.Duration, clusterCfg IClusterConfiguration) (*MulitClient, error) {
	mulitCli := &MulitClient{
		reBuildInterval: autoRbTime,
		stopCh:          make(chan struct{}, 0),
		ClusterCfg:      clusterCfg,
		ClusterCliMap:   map[string]*Client{},
	}

	clsList, err := mulitCli.ClusterCfg.GetAll()
	if err != nil {
		return nil, fmt.Errorf("get all cluster info failed:%+v", err)
	}

	for _, clsInfo := range clsList {
		cli, err := buildClient(clsInfo, mulitCli.ClusterCfg.GetOptions()...)
		if err != nil {
			return nil, err
		}
		mulitCli.ClusterCliMap[clsInfo.GetName()] = cli
	}

	return mulitCli, nil
}

func buildClient(clsInfo IClusterInfo, options ...Option) (*Client, error) {
	opts := []Option{}
	opts = append(opts, WithClusterName(clsInfo.GetName()))
	opts = append(opts, WithKubeConfig(clsInfo.GetKubeConfig()))
	opts = append(opts, options...)

	return NewClient(opts...)
}

// AddEventHandler add event with mulitclient
func (mc *MulitClient) AddEventHandler(handler cache.ResourceEventHandler, obj client.Object) error {
	var err error
	for name, cli := range mc.ClusterCliMap {
		err = cli.AddEventHandler(obj, handler)
		if err != nil {
			return fmt.Errorf("cluster [%s] AddEventHandler failed:%+v", name, err)
		}
	}
	return nil
}

// TriggerObjSync only trigger informer sync obj
func (mc *MulitClient) TriggerObjSync(obj client.Object) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for name, cli := range mc.ClusterCliMap {
		_, err = cli.GetInformerWithObj(obj)
		if err != nil {
			return fmt.Errorf("cluster [%s] TriggerObjSync failed:%+v", name, err)
		}
	}
	return nil
}

// SetIndexField set informer indexfield
func (mc *MulitClient) SetIndexField(obj client.Object, field string, extractValue client.IndexerFunc) error {
	mc.l.Lock()
	defer mc.l.Unlock()

	var err error
	for name, cli := range mc.ClusterCliMap {
		err = cli.CtrRtManager.GetFieldIndexer().IndexField(context.TODO(), obj, field, extractValue)
		if err != nil {
			return fmt.Errorf("cluster [%s] SetIndexField [%s] failed:%+v", name, field, err)
		}
	}
	return nil
}

// GetConnectedWithName get cluster with name and cluster is healthy.
func (mc *MulitClient) GetConnectedWithName(name string) (*Client, error) {
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
func (mc *MulitClient) GetReadyWithName(name string) (*Client, error) {
	cli, err := mc.GetWithName(name)
	if err != nil {
		return nil, err
	}
	if cli.ConnectStatus == Connected && cli.HasSynced() {
		return cli, nil
	}
	return nil, fmt.Errorf("cluster [%s] not connected apiserver or not ready", name)
}

// GetWithName get cluster with name.
func (mc *MulitClient) GetWithName(name string) (*Client, error) {
	mc.l.Lock()
	defer mc.l.Unlock()

	cli, ok := mc.ClusterCliMap[name]
	if !ok {
		return nil, fmt.Errorf("cluster [%s] not found, maybe not registry", name)
	}
	return cli, nil
}

// GetAllConnected get all cluster when cluster is connected.
func (mc *MulitClient) GetAllConnected() []*Client {
	mc.l.Lock()
	defer mc.l.Unlock()

	cliList := make([]*Client, 0, len(mc.ClusterCliMap))
	for _, cli := range mc.ClusterCliMap {
		if cli.ConnectStatus == Connected {
			cliList = append(cliList, cli)
		}
	}
	return cliList
}

// GetAllReady get all cluster when cluster is connected and informer has synced
func (mc *MulitClient) GetAllReady() []*Client {
	mc.l.Lock()
	defer mc.l.Unlock()

	cliList := make([]*Client, 0, len(mc.ClusterCliMap))
	for _, cli := range mc.ClusterCliMap {
		if cli.ConnectStatus == Connected && cli.HasSynced() {
			cliList = append(cliList, cli)
		}
	}
	return cliList
}

// GetAll get all cluster.
func (mc *MulitClient) GetAll() []*Client {
	mc.l.Lock()
	defer mc.l.Unlock()

	cliList := make([]*Client, 0, len(mc.ClusterCliMap))
	for _, cli := range mc.ClusterCliMap {
		cliList = append(cliList, cli)
	}
	return cliList
}

// HasSynced return all cluster has synced
func (mc *MulitClient) HasSynced() bool {
	mc.l.Lock()
	defer mc.l.Unlock()

	for _, cli := range mc.ClusterCliMap {
		if !cli.HasSynced() {
			return false
		}
	}
	return true
}

// Start start mulitclient
func (mc *MulitClient) Start(ctx context.Context) error {
	if mc.started {
		return errors.New("not restart mulitclient")
	}
	mc.l.Lock()
	defer mc.l.Unlock()

	mc.started = true
	go mc.autoRebuild()

	mc.ctx = ctx
	var err error
	for _, cli := range mc.ClusterCliMap {
		err = startClient(mc.ctx, cli, mc.BeforeStartFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

func startClient(ctx context.Context, cli *Client, beforeFunc BeforeStartFunc) error {
	if beforeFunc != nil {
		err := beforeFunc(cli)
		if err != nil {
			return fmt.Errorf("invoke cluster [%s] BeforeStartFunc failed:%+v", cli.GetName(), err)
		}
	}

	go func(cli *Client) {
		er := cli.Start(ctx)
		if er != nil {
			klog.Errorf("start cluster [%s] failed:%+v", cli.GetName(), er)
		}
	}(cli)
	return nil
}

// Stop mulitclient
func (mc *MulitClient) Stop() {
	close(mc.stopCh)

	wg := &sync.WaitGroup{}
	for _, cli := range mc.ClusterCliMap {
		wg.Add(1)
		go func(cli *Client) {
			cli.Stop()
			wg.Done()
		}(cli)
	}
	wg.Wait()
	return
}

func (mc *MulitClient) autoRebuild() {
	if mc.reBuildInterval <= 0 {
		return
	}

	for {
		select {
		case <-mc.stopCh:
			return
		case <-time.After(mc.reBuildInterval):
			mc.Rebuild()
		}
	}
}

// Rebuild rebuild with cluster info
func (mc *MulitClient) Rebuild() {
	if !mc.started {
		return
	}
	mc.l.Lock()
	defer mc.l.Unlock()

	newClsList, err := mc.ClusterCfg.GetAll()
	if err != nil {
		klog.Errorf("get all cluster info failed:%+v", err)
		return
	}

	newCliMap := make(map[string]*Client, len(newClsList))
	// add and check new cluster
	for _, newcls := range newClsList {
		// get old client info
		oldcli, exist := mc.ClusterCliMap[newcls.GetName()]
		if exist && oldcli.kubeconfig == newcls.GetKubeConfig() {
			// if kubeconfig not modify
			newCliMap[oldcli.GetName()] = oldcli
			continue
		}

		// build new client
		cli, err := buildClient(newcls, mc.ClusterCfg.GetOptions()...)
		if err != nil {
			klog.Error(err)
			continue
		}

		// start new client
		err = startClient(mc.ctx, cli, mc.BeforeStartFunc)
		if err != nil {
			klog.Error(err)
			return
		}

		if exist {
			// kubeconfig modify, should stop old client
			oldcli.Stop()
		}

		newCliMap[cli.GetName()] = cli
		klog.Infof("auto add cluster %s", cli.GetName())
	}

	// remove unexpect cluster
	for name, oldcli := range mc.ClusterCliMap {
		if _, ok := newCliMap[name]; !ok {
			// not exist, should stop
			go func(cli *Client) {
				cli.Stop()
			}(oldcli)
		}
	}

	mc.ClusterCliMap = newCliMap
}
