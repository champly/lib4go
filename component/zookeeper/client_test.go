package zookeeper

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"k8s.io/klog/v2"
)

var (
	servers = []string{
		"10.49.18.230",
	}
)

func TestNewClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	client, err := NewClient(ctx, servers, time.Second*5)
	if err != nil {
		t.Error(err)
		return
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
		return
	}
	t.Log(client.CreatePersistentNode("/a/a1", "永久节点"))
	t.Log(client.CreateTempNode("/a/a2", "临时节点"))
	t.Log(client.CreatePersistentSeqNode("/a/a3", "永久有序节点"))
	t.Log(client.CreateTempSeqNode("/a/a4", "临时序节点"))

	t.Log(client.CreatePersistentNode("/abc/a", ""))

	err = client.WatchValue("/abc/a", func(path, data string) bool {
		klog.Infoln("------------> data change:", path, data)
		return false
	})
	if err != nil {
		klog.Errorln(err)
	}

	err = client.WatchChildren("/abc/a", func(path string, children []string) bool {
		klog.Infoln("------------> change change:", path, children)
		return false
	})
	if err != nil {
		klog.Errorln(err)
	}
}

func TestWatchChildren(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	client, err := NewClient(ctx, servers, time.Second*5)
	if err != nil {
		t.Error(err)
	}

	path := "/watchchildren"
	client.Delete(path)

	t.Log("start create base dir")
	err = client.CreatePersistentNode(path, "")
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err := client.Delete(path)
		if err != nil {
			t.Error(err)
		}
	}()

	childrenPrefix := "a"
	num := 30
	children := make([]string, 0, num)

	for i := 0; i < num; i++ {
		children = append(children, fmt.Sprintf("%s%d", childrenPrefix, i))
	}

	actualChildren := []string{}
	go func() {
		err := client.WatchChildren(path, func(path string, children []string) bool {
			actualChildren = children
			return true
		})
		if err != nil {
			t.Error(err)
			return
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(num)
	for _, chil := range children {
		go func(chil string) {
			defer wg.Done()
			err = client.CreatePersistentNode(path+"/"+chil, "")
			if err != nil {
				t.Error(err)
				return
			}
		}(chil)
	}
	wg.Wait()
	time.Sleep(time.Second * 2)

	if !CompareSlice(children, actualChildren) {
		t.Errorf("children is not equal actualChildren:%+v\n%+v", actualChildren, children)
		return
	}
	t.Log("children equal actualChildren")
}
