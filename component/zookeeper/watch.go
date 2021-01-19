package zookeeper

import (
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

var (
	baseWorkInterval = time.Millisecond * 200
)

// HandlerValue handler watch value event
type HandlerValue func(path string, data string) (continueWatch bool)

// HandlerChildren handler watch children event
type HandlerChildren func(path string, children []string) (continueWatch bool)

type cb interface {
	Reconnect() (isContinue bool, err error)
	Event(e zk.Event) (isContinue bool, err error)
}

type basework struct {
	ch        <-chan zk.Event
	conn      *complexConn
	connected bool
	cb        cb
}

func (bw *basework) watch() error {
	var (
		isContinue bool
		err        error
		t          = time.NewTicker(baseWorkInterval)
	)
	defer t.Stop()

	for {
		select {
		case <-bw.conn.ctx.Done():
			return nil
		default:
		}

		if !bw.conn.isConnect() {
			bw.connected = false
			<-t.C
			continue
		}

		if !bw.connected {
			// client connected, should mark basework connected
			isContinue, err = bw.cb.Reconnect()
			if err != nil || !isContinue {
				return err
			}
			bw.connected = true
		}

		select {
		case <-bw.conn.ctx.Done():
			return nil
		case e, ok := <-bw.ch:
			if !ok {
				return nil
			}
			isContinue, err = bw.cb.Event(e)
			if err != nil || !isContinue {
				return nil
			}
		case <-t.C:
		}
	}
}

type workValue struct {
	*basework

	path    string
	value   string
	version int32
	handler HandlerValue
}

func (wv *workValue) Reconnect() (bool, error) {
	data, stat, ch, err := wv.conn.GetW(wv.path)
	if err != nil {
		return Stop, fmt.Errorf("reconnect watch value path [%s] fail:%s", wv.path, err)
	}
	wv.ch = ch
	if wv.version != stat.Version {
		wv.value = string(data)
		return wv.handler(wv.path, wv.value), nil
	}

	return Continue, nil
}

func (wv *workValue) Event(e zk.Event) (bool, error) {
	switch e.Type {

	case zk.EventNodeDataChanged:
		data, stat, ch, err := wv.conn.GetW(wv.path)
		if err != nil {
			return Stop, fmt.Errorf("rewatch value path [%s] fail:%s", wv.path, err)
		}
		wv.value = string(data)
		wv.ch = ch
		wv.version = stat.Version
		return wv.handler(wv.path, wv.value), nil

	default:
		data, stat, ch, err := wv.conn.GetW(wv.path)
		if err != nil {
			if wv.conn.isConnect() {
				return Stop, fmt.Errorf("rewatch value path [%s] fail:%s", wv.path, err)
			}
			// !import event connect, but conn is closed, should wait reconnected
			return Continue, nil
		}

		wv.ch = ch
		if wv.version != stat.Version {
			wv.value = string(data)
			wv.version = stat.Version
			return wv.handler(wv.path, wv.value), nil
		}

		return Continue, nil
	}
}

// WatchValue watch value change
func (zc *ZkClient) WatchValue(path string, handler HandlerValue) error {
	if b, e := zc.IsExists(path); !b || e != nil {
		return fmt.Errorf("path [%s] not exists:%v", path, e)
	}

	conn, err := zc.getComplexConn()
	if err != nil {
		return err
	}
	data, stat, ch, err := conn.GetW(path)
	if err != nil {
		return fmt.Errorf("path [%s] getw error:%+v", path, err)
	}
	// first invoke handler
	if !handler(path, string(data)) {
		return nil
	}

	wv := &workValue{
		basework: &basework{
			ch:        ch,
			conn:      conn,
			connected: conn.isConnect(),
		},
		path:    path,
		value:   string(data),
		version: stat.Version,
		handler: handler,
	}
	wv.cb = wv

	return wv.watch()
}

type workChildren struct {
	*basework

	path     string
	children []string
	cversion int32
	handler  HandlerChildren
}

func (wc *workChildren) Reconnect() (bool, error) {
	children, stat, ch, err := wc.conn.ChildrenW(wc.path)
	if err != nil {
		return Stop, fmt.Errorf("reconnect watch children path [%s] fail:%s", wc.path, err)
	}
	wc.ch = ch
	if wc.cversion != stat.Cversion {
		wc.children = children
		wc.cversion = stat.Cversion
		return wc.handler(wc.path, wc.children), nil
	}
	return Continue, nil
}

func (wc *workChildren) Event(e zk.Event) (bool, error) {
	switch e.Type {

	case zk.EventNodeChildrenChanged:
		children, stat, ch, err := wc.conn.ChildrenW(wc.path)
		if err != nil {
			return Stop, fmt.Errorf("rewatch children path [%s] fail:%s", wc.path, err)
		}
		wc.children = children
		wc.ch = ch
		wc.cversion = stat.Cversion
		return wc.handler(wc.path, wc.children), nil

	default:
		children, stat, ch, err := wc.conn.ChildrenW(wc.path)
		if err != nil {
			if wc.conn.isConnect() {
				return Stop, fmt.Errorf("rewatch children path [%s] fail:%s", wc.path, err)
			}
			// !import event connect, but conn is closed, should wait reconnected
			return Continue, nil
		}

		wc.ch = ch
		if wc.cversion != stat.Version {
			wc.children = children
			return wc.handler(wc.path, wc.children), nil
		}
		return Continue, nil
	}
}

// WatchChildren watch children
func (zc *ZkClient) WatchChildren(path string, handler HandlerChildren) error {
	if b, e := zc.IsExists(path); !b || e != nil {
		return fmt.Errorf("path [%s] not exists:%v", path, e)
	}

	conn, err := zc.getComplexConn()
	if err != nil {
		return err
	}
	children, stat, ch, err := conn.ChildrenW(path)
	if err != nil {
		return fmt.Errorf("path [%s] childrenw error:%+v", path, err)
	}
	// first invoke handler
	if !handler(path, children) {
		return nil
	}

	wc := &workChildren{
		basework: &basework{
			ch:        ch,
			conn:      conn,
			connected: conn.isConnect(),
		},
		path:     path,
		children: children,
		cversion: stat.Cversion,
		handler:  handler,
	}
	wc.cb = wc

	return wc.watch()
}
