package zookeeper

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"k8s.io/klog"
)

// HandlerValue handler watch value event
type HandlerValue func(path string, data string, isEnd bool)

type cb interface {
	Reconnect() bool
	Event(e zk.Event) bool
}

type basework struct {
	ch        <-chan zk.Event
	stopCh    <-chan struct{}
	zcli      *ZkClient
	connected bool
	cb        cb
}

func (bw *basework) run() {
	t := time.NewTimer(time.Microsecond * 100)
	defer t.Stop()

	for {
		if !bw.zcli.IsConnect() {
			bw.connected = false
			<-t.C
			continue
		}

		if !bw.connected {
			if !bw.cb.Reconnect() {
				return
			}
		}

		select {
		case <-bw.stopCh:
			return
		case e, ok := <-bw.ch:
			if !ok {
				return
			}
			if !bw.cb.Event(e) {
				return
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

func (wv *workValue) Reconnect() bool {
	data, _, ch, err := wv.zcli.conn.GetW(wv.path)
	if err != nil {
		klog.Errorf("reconnect watch value path [%s] fail:%s", wv.path, err)
		wv.handler(wv.path, "", true)
		return Stop
	}
	wv.ch = ch
	if !strings.EqualFold(wv.value, string(data)) {
		wv.value = string(data)
		wv.handler(wv.path, wv.value, false)
	}

	return Continue
}

func (wv *workValue) Event(e zk.Event) bool {
	switch e.Type {

	case zk.EventNodeDataChanged:
		data, _, ch, err := wv.zcli.conn.GetW(wv.path)
		if err != nil {
			klog.Errorf("rewatch value path [%s] fail:%s", wv.path, err)
			wv.handler(wv.path, "", true)
			return Stop
		}
		wv.value = string(data)
		wv.ch = ch
		wv.handler(wv.path, wv.value, false)
		return Continue

	default:
		data, _, ch, err := wv.zcli.conn.GetW(wv.path)
		if err != nil {
			klog.Errorf("rewatch value path [%s] fail:%s", wv.path, err)
			if wv.zcli.IsConnect() {
				wv.handler(wv.path, "", true)
				return Stop
			}
			// !import event connect, but conn is closed, should wait reconnected
			return Continue
		}

		wv.ch = ch
		if !strings.EqualFold(wv.value, string(data)) {
			wv.value = string(data)
			wv.handler(wv.path, wv.value, false)
		}

		return Continue
	}
}

// WatchValue watch value change
func (z *ZkClient) WatchValue(path string, handler HandlerValue) (err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	if b, e := z.IsExists(path); !b || e != nil {
		return fmt.Errorf("path [%s] not exists:%v", path, e)
	}

	data, _, ch, err := z.conn.GetW(path)
	if err != nil {
		return
	}

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

	go wv.run()

	return nil
}

// HandlerChildren handler watch children event
type HandlerChildren func(path string, children []string, isEnd bool)

type workChildren struct {
	*basework

	path     string
	children []string
	handler  HandlerChildren
}

func (wc *workChildren) Reconnect() bool {
	children, _, ch, err := wc.zcli.conn.ChildrenW(wc.path)
	if err != nil {
		klog.Errorf("reconnect watch value path [%s] fail:%s", wc.path, err)
		wc.handler(wc.path, nil, true)
		return Stop
	}
	wc.ch = ch
	if !CompareSlice(wc.children, children) {
		wc.children = children
		wc.handler(wc.path, wc.children, false)
	}
	return Continue
}

func (wc *workChildren) Event(e zk.Event) bool {
	switch e.Type {

	case zk.EventNodeChildrenChanged:
		children, _, ch, err := wc.zcli.conn.ChildrenW(wc.path)
		if err != nil {
			klog.Errorf("rewatch value path [%s] fail:%s", wc.path, err)
			wc.handler(wc.path, nil, true)
			return Stop
		}
		wc.children = children
		wc.ch = ch
		wc.handler(wc.path, wc.children, false)
		return Continue

	default:
		children, _, ch, err := wc.zcli.conn.ChildrenW(wc.path)
		if err != nil {
			klog.Errorf("rewatch value path [%s] fail:%s", wc.path, err)
			if wc.zcli.IsConnect() {
				wc.handler(wc.path, nil, true)
				return Stop
			}
			// !import event connect, but conn is closed, should wait reconnected
			return Continue
		}

		wc.ch = ch
		if !CompareSlice(wc.children, children) {
			wc.children = children
			wc.handler(wc.path, wc.children, false)
		}
		return Continue
	}
}

// WatchChildren watch children
func (z *ZkClient) WatchChildren(path string, handler HandlerChildren) (err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	if b, e := z.IsExists(path); !b || e != nil {
		return fmt.Errorf("path [%s] not exists:%v", path, e)
	}

	children, _, ch, err := z.conn.ChildrenW(path)
	if err != nil {
		return
	}

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

	go wc.run()

	return nil
}
