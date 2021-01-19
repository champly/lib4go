package zookeeper

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

// HandlerValue handler watch value event
type HandlerValue func(path string, data string)

// HandlerChildren handler watch children event
type HandlerChildren func(path string, children []string)

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
		t          = time.NewTimer(time.Microsecond * 100)
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
	handler HandlerValue
}

func (wv *workValue) Reconnect() (bool, error) {
	data, _, ch, err := wv.conn.GetW(wv.path)
	if err != nil {
		return Stop, fmt.Errorf("reconnect watch value path [%s] fail:%s", wv.path, err)
	}
	wv.ch = ch
	if !strings.EqualFold(wv.value, string(data)) {
		wv.value = string(data)
		wv.handler(wv.path, wv.value)
	}

	return Continue, nil
}

func (wv *workValue) Event(e zk.Event) (bool, error) {
	switch e.Type {

	case zk.EventNodeDataChanged:
		data, _, ch, err := wv.conn.GetW(wv.path)
		if err != nil {
			return Stop, fmt.Errorf("rewatch value path [%s] fail:%s", wv.path, err)
		}
		wv.value = string(data)
		wv.ch = ch
		wv.handler(wv.path, wv.value)
		return Continue, nil

	default:
		data, _, ch, err := wv.conn.GetW(wv.path)
		if err != nil {
			if wv.conn.isConnect() {
				return Stop, fmt.Errorf("rewatch value path [%s] fail:%s", wv.path, err)
			}
			// !import event connect, but conn is closed, should wait reconnected
			return Continue, nil
		}

		wv.ch = ch
		if !strings.EqualFold(wv.value, string(data)) {
			wv.value = string(data)
			wv.handler(wv.path, wv.value)
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
	data, _, ch, err := conn.GetW(path)
	if err != nil {
		return fmt.Errorf("path [%s] getw error:%+v", path, err)
	}
	// first invoke handler
	handler(path, string(data))

	wv := &workValue{
		basework: &basework{
			ch:        ch,
			conn:      conn,
			connected: conn.isConnect(),
		},
		path:    path,
		value:   string(data),
		handler: handler,
	}
	wv.cb = wv

	return wv.watch()
}

type workChildren struct {
	*basework

	path     string
	children []string
	handler  HandlerChildren
}

func (wc *workChildren) Reconnect() (bool, error) {
	children, _, ch, err := wc.conn.ChildrenW(wc.path)
	if err != nil {
		return Stop, fmt.Errorf("reconnect watch children path [%s] fail:%s", wc.path, err)
	}
	wc.ch = ch
	if !CompareSlice(wc.children, children) {
		wc.children = children
		wc.handler(wc.path, wc.children)
	}
	return Continue, nil
}

func (wc *workChildren) Event(e zk.Event) (bool, error) {
	switch e.Type {

	case zk.EventNodeChildrenChanged:
		children, _, ch, err := wc.conn.ChildrenW(wc.path)
		if err != nil {
			return Stop, fmt.Errorf("rewatch children path [%s] fail:%s", wc.path, err)
		}
		wc.children = children
		wc.ch = ch
		wc.handler(wc.path, wc.children)
		return Continue, nil

	default:
		children, _, ch, err := wc.conn.ChildrenW(wc.path)
		if err != nil {
			if wc.conn.isConnect() {
				return Stop, fmt.Errorf("rewatch children path [%s] fail:%s", wc.path, err)
			}
			// !import event connect, but conn is closed, should wait reconnected
			return Continue, nil
		}

		wc.ch = ch
		if !CompareSlice(wc.children, children) {
			wc.children = children
			wc.handler(wc.path, wc.children)
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
	children, _, ch, err := conn.ChildrenW(path)
	if err != nil {
		return fmt.Errorf("path [%s] childrenw error:%+v", path, err)
	}
	// first invoke handler
	handler(path, children)

	wc := &workChildren{
		basework: &basework{
			ch:        ch,
			conn:      conn,
			connected: conn.isConnect(),
		},
		path:     path,
		children: children,
		handler:  handler,
	}
	wc.cb = wc

	return wc.watch()
}
