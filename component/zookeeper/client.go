package zookeeper

import (
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

// ZkClient zk client struct
type ZkClient struct {
	*option

	servers []string
	timeout time.Duration

	stopCh <-chan struct{}

	conn      *zk.Conn
	eventChan <-chan zk.Event
	isConnect bool
}

// New new zk client
func New(servers []string, timeout time.Duration, opts ...Option) (*ZkClient, error) {
	client := &ZkClient{
		option:  DefaultOption(),
		servers: servers,
		timeout: timeout,
		stopCh:  make(<-chan struct{}),
	}

	for _, opt := range opts {
		opt(client.option)
	}

	if err := client.connect(); err != nil {
		return nil, err
	}

	err := WarpperTimeout(func() {
		for !client.IsConnect() {
			time.Sleep(time.Millisecond * 10)
		}
	}, client.connectTimeout)
	if err != nil {
		return nil, errors.Wrapf(err, "connect server {%+v} fail:%v", client.servers, err)
	}

	return client, nil
}

// Start connect zk
func (z *ZkClient) Start(ch <-chan struct{}) error {
	<-ch
	return z.stop()
}

// IsConnect return is connect zk server
func (z *ZkClient) IsConnect() bool {
	return z.isConnect
}

// connect zk service
func (z *ZkClient) connect() error {
	if z.conn == nil {
		conn, eventChan, err := zk.Connect(z.servers, z.timeout)
		if err != nil {
			return errors.Wrap(err, "connect zk servers fail")
		}
		z.conn = conn
		z.eventChan = eventChan
		go z.eventReceive()
	}

	return nil
}

func (z *ZkClient) eventReceive() {
	for {
		select {
		case <-z.stopCh:
			return
		case v, ok := <-z.eventChan:
			if !ok {
				z.isConnect = false
				return
			}

			switch v.State {
			case zk.StateUnknown:
				klog.Infof("unknow state:%+v", v)
			case zk.StateDisconnected:
				klog.Infof("zk {%s} disconnected", z.servers)
				z.isConnect = false
			case zk.StateConnecting:
				klog.Infof("zk {%s} is connecting", z.servers)
				z.isConnect = false
			case zk.StateAuthFailed:
				klog.Infof("zk {%s} auth failed", z.servers)
				z.stop()
			case zk.StateConnectedReadOnly:
				klog.Infof("zk {%s} connected but read only", z.servers)
				z.isConnect = true
			case zk.StateSaslAuthenticated:
				klog.Infof("zk {%s} sas authenticated", z.servers)
			case zk.StateExpired:
				klog.Infof("zk {%s} state expired", z.servers)
				z.isConnect = false
			case zk.StateConnected:
				klog.Infof("zk {%s} connected", z.servers)
				z.isConnect = true
			case zk.StateHasSession:
				klog.Infof("zk {%s} has session", z.servers)
				z.isConnect = true
			default:
				klog.Infof("undefine state:%+v", v)
			}
		}
	}
}

// stop zk client
func (z *ZkClient) stop() error {
	z.isConnect = false
	if z.conn != nil {
		z.conn.Close()
	}
	return nil
}

// // Use unit test
// func (z *ZkClient) Close() {
// 	z.conn.Close()
// 	z.isConnect = false
// 	z.conn = nil
// }

// func (z *ZkClient) Connect() {
// 	z.connect()
// }
