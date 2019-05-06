package http

import (
	"net/http"
	"testing"
)

type benchCB struct {
}

func (b *benchCB) Before(req *http.Request) {
}

func (b *benchCB) After(req *http.Request, resp *http.Response, err error) {
}

func BenchmarkNewClientWithCb(b *testing.B) {
	bcb := &benchCB{}
	for i := 0; i < b.N; i++ {
		client := NewClientWithCb(bcb)
		client.Get("http://localhost:9090/test")
	}
}
