package http

import (
	"net/http"
	"testing"
)

type benchCB struct {
}

func (b *benchCB) Before(req *http.Request) {
}

func (b *benchCB) After(resp *http.Response, err error) {
}

func BenchmarkNewClientWithCb(b *testing.B) {
	bcb := &benchCB{}
	for i := 0; i < b.N; i++ {
		client := NewClientWithCb(bcb)
		client.Get("http://www.baidu.com")
	}
}
