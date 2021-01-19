package zookeeper

import (
	"context"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// ZkClient zk client struct
type ZkClient struct {
	*option
	index    int
	connPool []*complexConn
	timeout  time.Duration
	l        sync.Mutex
}

type complexConn struct {
	*zk.Conn
	ctx       context.Context
	servers   []string
	eventCh   <-chan zk.Event
	connected bool
}

// NewClient new zk client
func NewClient(ctx context.Context, servers []string, timeout time.Duration, opts ...Option) (*ZkClient, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	client := &ZkClient{
		option:  DefaultOption(),
		timeout: timeout,
	}

	for _, opt := range opts {
		opt(client.option)
	}

	if err := client.connect(ctx, servers); err != nil {
		return nil, err
	}

	return client, nil
}

// connect zk service
func (zc *ZkClient) connect(ctx context.Context, servers []string) error {
	zc.connPool = make([]*complexConn, 0, zc.connectNum)
	for i := 0; i < zc.connectNum; i++ {
		conn, eventCh, err := zk.Connect(servers, zc.timeout)
		if err != nil {
			return errors.Wrap(err, "connect zk servers fail")
		}
		cc := &complexConn{
			Conn:    conn,
			ctx:     ctx,
			servers: servers,
			eventCh: eventCh,
		}
		go cc.eventReceive()

		err = WarpperTimeout(func() {
			for !cc.isConnect() {
				time.Sleep(time.Millisecond * 10)
			}
		}, zc.connectTimeout)
		if err != nil {
			return errors.Wrapf(err, "connect server {%+v} fail:%v", servers, err)
		}

		zc.connPool = append(zc.connPool, cc)
	}
	return nil
}

func (zc *ZkClient) getComplexConn() (*complexConn, error) {
	zc.l.Lock()
	defer zc.l.Unlock()

	conn := zc.connPool[zc.index]
	zc.index = (zc.index + 1) % zc.connectNum
	if !conn.isConnect() {
		return nil, ErrClientDisConnect
	}
	return conn, nil
}

func (cc *complexConn) eventReceive() {
	for {
		select {
		case <-cc.ctx.Done():
			cc.close()
			return
		case v, ok := <-cc.eventCh:
			if !ok {
				cc.connected = false
				return
			}

			switch v.State {
			case zk.StateUnknown:
				klog.V(5).Infof("unknow state:%+v", v)
			case zk.StateDisconnected:
				klog.V(5).Infof("zk {%s} disconnected", cc.servers)
				cc.connected = false
			case zk.StateConnecting:
				klog.V(5).Infof("zk {%s} is connecting", cc.servers)
				cc.connected = false
			case zk.StateAuthFailed:
				klog.V(5).Infof("zk {%s} auth failed", cc.servers)
				cc.close()
			case zk.StateConnectedReadOnly:
				klog.V(5).Infof("zk {%s} connected but read only", cc.servers)
				cc.connected = true
			case zk.StateSaslAuthenticated:
				klog.V(5).Infof("zk {%s} sas authenticated", cc.servers)
			case zk.StateExpired:
				klog.V(5).Infof("zk {%s} state expired", cc.servers)
				cc.connected = false
			case zk.StateConnected:
				klog.V(5).Infof("zk {%s} connected", cc.servers)
				cc.connected = true
			case zk.StateHasSession:
				klog.V(5).Infof("zk {%s} has session", cc.servers)
				cc.connected = true
			default:
				klog.V(5).Infof("undefine state:%+v", v)
			}
		}
	}
}

// isConnect return is connect zk server
func (cc *complexConn) isConnect() bool {
	return cc.connected
}

// stop zk client
func (cc *complexConn) close() {
	cc.connected = false
	if cc.Conn != nil {
		cc.Conn.Close()
	}
	klog.Warningf("zk client %+v stoped", cc.servers)
}
