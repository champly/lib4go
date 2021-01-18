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
	stopCh    <-chan struct{}
	zcli      *ZkClient
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
		case <-bw.zcli.stopCh:
			return nil
		default:
		}

		if !bw.zcli.IsConnect() {
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
		case <-bw.zcli.stopCh:
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
	data, _, ch, err := wv.zcli.conn.GetW(wv.path)
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
		data, _, ch, err := wv.zcli.conn.GetW(wv.path)
		if err != nil {
			return Stop, fmt.Errorf("rewatch value path [%s] fail:%s", wv.path, err)
		}
		wv.value = string(data)
		wv.ch = ch
		wv.handler(wv.path, wv.value)
		return Continue, nil

	default:
		data, _, ch, err := wv.zcli.conn.GetW(wv.path)
		if err != nil {
			if wv.zcli.IsConnect() {
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
func (z *ZkClient) WatchValue(path string, handler HandlerValue) error {
	if !z.isConnect {
		return ErrClientDisConnect
	}

	if b, e := z.IsExists(path); !b || e != nil {
		return fmt.Errorf("path [%s] not exists:%v", path, e)
	}

	data, _, ch, err := z.conn.GetW(path)
	if err != nil {
		return fmt.Errorf("path [%s] getw error:%+v", path, err)
	}
	// first invoke handler
	handler(path, string(data))

	wv := &workValue{
		basework: &basework{
			ch:        ch,
			stopCh:    z.stopCh,
			zcli:      z,
			connected: z.IsConnect(),
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
	children, _, ch, err := wc.zcli.conn.ChildrenW(wc.path)
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
		children, _, ch, err := wc.zcli.conn.ChildrenW(wc.path)
		if err != nil {
			return Stop, fmt.Errorf("rewatch children path [%s] fail:%s", wc.path, err)
		}
		wc.children = children
		wc.ch = ch
		wc.handler(wc.path, wc.children)
		return Continue, nil

	default:
		children, _, ch, err := wc.zcli.conn.ChildrenW(wc.path)
		if err != nil {
			if wc.zcli.IsConnect() {
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
func (z *ZkClient) WatchChildren(path string, handler HandlerChildren) error {
	if !z.isConnect {
		return ErrClientDisConnect
	}

	if b, e := z.IsExists(path); !b || e != nil {
		return fmt.Errorf("path [%s] not exists:%v", path, e)
	}

	children, _, ch, err := z.conn.ChildrenW(path)
	if err != nil {
		return fmt.Errorf("path [%s] childrenw error:%+v", path, err)
	}
	// first invoke handler
	handler(path, children)

	wc := &workChildren{
		basework: &basework{
			ch:        ch,
			stopCh:    z.stopCh,
			zcli:      z,
			connected: z.IsConnect(),
		},
		path:     path,
		children: children,
		handler:  handler,
	}
	wc.cb = wc

	return wc.watch()
}
