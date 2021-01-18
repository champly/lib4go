package zookeeper

import (
	"testing"
	"time"

	"github.com/champly/lib4go/signal"
	"k8s.io/klog/v2"
)

func TestNewClient(t *testing.T) {
	zk, err := New([]string{"127.0.0.1:2181"}, time.Second*5)
	if err != nil {
		t.Error(err)
	}

	go func() {
		if err := zk.Start(signal.SetupSignalHandler()); err != nil {
			t.Error(err)
		}
	}()

	for !zk.IsConnect() {
		time.Sleep(time.Second * 1)
	}

	path := "/a/b/c/d"
	value := "中文测试"
	if err = zk.CreatePersistentNode(path, value); err != nil {
		t.Error(err)
	}

	t.Log(zk.GetValue(path))

	// if err = zk.Delete("/a"); err != nil {
	// 	t.Error(err)
	// }

	// if err = zk.Delete("/dubbo"); err != nil {
	// 	t.Error(err)
	// }

	if err = zk.CreateTempNode("/a", "临时节点"); err != nil {
		t.Error(err)
	}

	// t.Log(zk.CreatePersistentNode("/a1", "永久节点"))
	// t.Log(zk.CreateTempNode("/a2", "临时节点"))
	// t.Log(zk.CreatePersistentSeqNode("/a3", "永久有序节点"))
	// t.Log(zk.CreateTempSeqNode("/a4", "临时序节点"))

	t.Log(zk.CreatePersistentNode("/abc/a", ""))

	done1 := make(chan struct{}, 0)
	go func() {
		err = zk.WatchValue("/abc/a", func(path, data string) {
			klog.Infoln("------------> data change:", path, data)
		})
		if err != nil {
			klog.Errorln(err)
		}
		close(done1)
	}()

	select {
	case <-time.After(time.Second * 1):
	case <-done1:
	}

	done2 := make(chan struct{}, 0)
	go func() {
		err = zk.WatchChildren("/abc/a", func(path string, children []string) {
			klog.Infoln("------------> change change:", path, children)
		})
		if err != nil {
			klog.Errorln(err)
		}
		close(done2)
	}()
	select {
	case <-time.After(time.Second * 1):
	case <-done2:
	}

	// time.Sleep(time.Second * 20)
	// zk.Close()
	// klog.Infoln("connected is close")
	// time.Sleep(time.Second * 20)
	// zk.Connect()
	// klog.Infoln("zk reconnected")

	// time.Sleep(time.Hour * 1)
}
