package zookeeper

import (
	"context"
	"testing"
	"time"

	"k8s.io/klog/v2"
)

func TestNewClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer func() {
		cancel()
	}()

	client, err := NewClient(ctx, []string{"10.49.18.230:2181"}, time.Second*5)
	if err != nil {
		t.Error(err)
	}

	path := "/a/b/c/d"
	value := "中文测试"
	if err = client.CreatePersistentNode(path, value); err != nil {
		t.Error(err)
	}

	t.Log(client.GetValue(path))

	defer func() {
		if err = client.Delete("/a"); err != nil {
			t.Error(err)
		}
	}()

	if err = client.CreateTempNode("/a", "临时节点"); err != nil {
		t.Error(err)
	}
	t.Log(client.CreatePersistentNode("/a/a1", "永久节点"))
	t.Log(client.CreateTempNode("/a/a2", "临时节点"))
	t.Log(client.CreatePersistentSeqNode("/a/a3", "永久有序节点"))
	t.Log(client.CreateTempSeqNode("/a/a4", "临时序节点"))

	t.Log(client.CreatePersistentNode("/abc/a", ""))

	done1 := make(chan struct{}, 0)
	go func() {
		err = client.WatchValue("/abc/a", func(path, data string) {
			klog.Infoln("------------> data change:", path, data)
		})
		if err != nil {
			klog.Errorln(err)
		}
		close(done1)
	}()

	select {
	case <-time.After(time.Second * 2):
	case <-done1:
	}

	done2 := make(chan struct{}, 0)
	go func() {
		err = client.WatchChildren("/abc/a", func(path string, children []string) {
			klog.Infoln("------------> change change:", path, children)
		})
		if err != nil {
			klog.Errorln(err)
		}
		close(done2)
	}()
	select {
	case <-time.After(time.Second * 2):
	case <-done2:
	}
}
