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

	// t.Log(zk.IsExists("/abc"))

	err = zk.WatchValue("/abc/a", func(path, data string, isEnd bool) {
		klog.Infoln("------------> data change:", path, data, isEnd)
	})
	if err != nil {
		klog.Errorln(err)
	}

	err = zk.WatchChildren("/abc/a", func(path string, children []string, isEnd bool) {
		klog.Infoln("------------> data change:", path, children, isEnd)
	})
	if err != nil {
		klog.Errorln(err)
	}

	// time.Sleep(time.Second * 20)
	// zk.Close()
	// klog.Infoln("connected is close")
	// time.Sleep(time.Second * 20)
	// zk.Connect()
	// klog.Infoln("zk reconnected")

	// time.Sleep(time.Hour * 1)
}
