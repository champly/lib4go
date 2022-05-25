package etcd3

import (
	"context"
	"errors"
	"path"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/pkg/transport"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

const (
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
	dialTimeout      = 20 * time.Second
)

type store struct {
	client     *clientv3.Client
	watcher    *watcher
	pathPrefix string
}

func New(opts ...Option) (s *store, err error) {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}

	client, err := buildClientV3(cfg)
	if err != nil {
		klog.Error("connect etcd failed:", err)
		return nil, err
	}
	klog.Info("connect success")

	return newStore(client, cfg.Prefix), nil
}

func buildClientV3(c *Config) (client *clientv3.Client, err error) {
	tlsInfo := transport.TLSInfo{
		KeyFile:       c.KeyFile,
		CertFile:      c.CertFile,
		TrustedCAFile: c.TrustedCAFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}

	// NOTE: Client relies on nil tlsConfig
	if len(c.CertFile) == 0 && len(c.TrustedCAFile) == 0 {
		tlsConfig = nil
	}

	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		// grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
		// grpc.WithUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		// grpc.WithStreamInterceptor(grpcprom.StreamClientInterceptor),
	}

	cfg := clientv3.Config{
		Endpoints:            c.ServerList,
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
		DialOptions:          dialOptions,
		TLS:                  tlsConfig,
	}

	return clientv3.New(cfg)
}

func newStore(client *clientv3.Client, prefix string) *store {
	return &store{
		client:     client,
		watcher:    newWatcher(client),
		pathPrefix: path.Join("/", prefix),
	}
}

func (s *store) Create(ctx context.Context, key, val string) error {
	key = path.Join(s.pathPrefix, key)

	var opts []clientv3.OpOption
	txnResp, err := s.client.KV.Txn(ctx).If(
		notFound(key),
	).Then(
		clientv3.OpPut(key, string(val), opts...),
	).Commit()
	if err != nil {
		return err
	}

	klog.Info(txnResp)
	if !txnResp.Succeeded {
		return errors.New("create key failed")
	}
	return nil
}

func (s *store) Get(ctx context.Context, key string) (val []byte, err error) {
	key = path.Join(s.pathPrefix, key)
	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, errors.New("not found data")
	}
	return resp.Kvs[0].Value, nil
}

func (s *store) Update(ctx context.Context, key string, val string) error {
	key = path.Join(s.pathPrefix, key)

	var opts []clientv3.OpOption
	txnResp, err := s.client.KV.Txn(ctx).If(
		clientv3.Compare(clientv3.ModRevision(key), "=", 5),
	).Then(
		clientv3.OpPut(key, val, opts...),
	).Else(
		clientv3.OpGet(key, opts...),
	).Commit()
	if err != nil {
		return err
	}
	klog.Info(txnResp)

	if !txnResp.Succeeded {
		return errors.New("resource not found or has modified")
	}
	return nil
}

func (s *store) Delete(ctx context.Context, key string) error {
	// key = path.Join(s.pathPrefix, key)

	panic("not implement")
}

func (s *store) List(ctx context.Context, key string) (out interface{}, err error) {
	// key = path.Join(s.pathPrefix, key)

	// s.client.KV.Get
	panic("not implement")
}

func notFound(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.ModRevision(key), "=", 0)
}
