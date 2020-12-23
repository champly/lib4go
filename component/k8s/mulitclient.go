package k8s

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MulitClient mulit cluster client obj
type MulitClient struct {
	l sync.Mutex

	BeforeStartFunc BeforeStartFunc
	ClusterCfg      IClusterConfiguration
	ClusterCliMap   map[string]*Client
}

// NewMulitClient build MulitClient
func NewMulitClient(clusterCfg IClusterConfiguration) (*MulitClient, error) {
	mulitCli := &MulitClient{
		ClusterCfg:    clusterCfg,
		ClusterCliMap: map[string]*Client{},
	}

	var err error
	for _, clsinfo := range mulitCli.ClusterCfg.GetAll() {
		opts := []Option{}
		opts = append(opts, WithClusterName(clsinfo.GetName()))
		opts = append(opts, WithKubeConfig(clsinfo.GetKubeConfig()))
		opts = append(opts, mulitCli.ClusterCfg.GetOptions()...)

		mulitCli.ClusterCliMap[clsinfo.GetName()], err = NewClient(opts...)
		if err != nil {
			return nil, err
		}
	}

	return mulitCli, nil
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
	for _, cli := range mc.ClusterCliMap {
		if !cli.HasSynced() {
			return false
		}
	}
	return true
}

// Start start mulitclient
func (mc *MulitClient) Start(ctx context.Context) error {
	var err error
	for name, cli := range mc.ClusterCliMap {
		err = mc.BeforeStartFunc(cli)
		if err != nil {
			return fmt.Errorf("invoke cluster [%s] BeforeStartFunc failed:%+v", name, err)
		}
		go func(cli *Client) {
			er := cli.Start(ctx)
			if er != nil {
				klog.Errorf("start cluster [%s] failed:%+v", cli.GetName(), er)
			}
		}(cli)
	}
	return nil
}

// Stop mulitclient
func (mc *MulitClient) Stop() {
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
