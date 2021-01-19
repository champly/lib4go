package zookeeper

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var (
	watchchildrenExceptionBaseDir = "/watchexcpetion"
	watchchildrenExceptionParent  = watchchildrenExceptionBaseDir + "/parent"
)

func TestWatchChildrenException(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	client, err := NewClient(ctx, servers, time.Second*5)
	if err != nil {
		t.Error(err)
		return
	}

	err = client.Delete(watchchildrenExceptionBaseDir)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = client.Delete(watchchildrenExceptionBaseDir)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	t.Run("watch children when parent delete", func(t *testing.T) {
		client.CreatePersistentNode(watchchildrenExceptionParent+"/demo", "")
		go func() {
			err := client.WatchChildren(fmt.Sprintf("%s/%s", watchchildrenExceptionParent, "demo"), func(path string, children []string) bool {
				t.Log(children)
				return false
			})
			if err != nil {
				t.Error(err)
				return
			}
		}()
		time.Sleep(time.Second * 1)
		client.Delete(watchchildrenExceptionParent)
	})

	t.Run("reconnected watch children", func(t *testing.T) {
		client.CreatePersistentNode(watchchildrenExceptionParent, "")
		go func() {
			err := client.WatchChildren(watchchildrenExceptionParent, func(path string, children []string) bool {
				t.Log(children)
				return true
			})
			if err != nil {
				t.Error(err)
				return
			}
		}()
		time.Sleep(time.Second * 1)
		client.connPool[0].connected = false
		time.Sleep(time.Second * 1)
		client.connPool[0].connected = true
		time.Sleep(time.Second * 1)
	})
}
