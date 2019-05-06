package http

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

type printCB struct {
	start time.Time
}

func (p *printCB) Before(req *http.Request) {
	p.start = time.Now()
	fmt.Println("host:", req.Host)
	fmt.Println("method:", req.Method)
}

func (p *printCB) After(req *http.Request, resp *http.Response, err error) {
	dur := time.Since(p.start)
	fmt.Println("sum use:", dur)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("request success")
	}
}

func TestNewClientWithCb(t *testing.T) {
	client := NewClientWithCb(&printCB{})
	client.Get("http://www.baidu.com")
}
