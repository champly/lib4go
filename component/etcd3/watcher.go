package etcd3

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type watcher struct {
	client *clientv3.Client
}

type watchChan struct {
	watcher *watcher
	ctx     context.Context
	cancel  context.CancelFunc
	errChan chan error
}

func newWatcher(client *clientv3.Client) *watcher {
	return &watcher{
		client: client,
	}
}
