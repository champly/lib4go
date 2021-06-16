package etcd3

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

/*
docker run --rm \
  -p 2379:2379 \
  -p 2380:2380 \
  --volume=/tmp/etcd-data:/etcd-data \
  --name etcd quay.io/coreos/etcd:v3.5.0 \
  /usr/local/bin/etcd \
  --data-dir=/etcd-data --name node1 \
  --advertise-client-urls http://0.0.0.0:2379 --listen-client-urls http://0.0.0.0:2379
*/

func TestGet(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Error(err)
		return
	}
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := s.Get(context.TODO(), "key")
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(data)
	}))
	http.Handle("/create", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := s.Create(context.TODO(), "key", time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte("ok"))
	}))
	http.Handle("/update", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := s.Update(context.TODO(), "key", time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte("ok"))
	}))

	if e := http.ListenAndServe(":8080", nil); e != nil {
		klog.Errorf("start http failed:", e)
	}
}
