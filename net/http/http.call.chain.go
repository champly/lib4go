package http

import (
	"fmt"
	"net/http"
	"time"
)

func NewClientWithCallChain() *http.Client {
	return &http.Client{
		Transport: newCallChainRoundTrip(http.DefaultTransport),
	}
}

type callChian struct {
	rt http.RoundTripper
}

func newCallChainRoundTrip(rt http.RoundTripper) http.RoundTripper {
	return &callChian{rt: rt}
}

func (c *callChian) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	startTime := time.Now()
	fmt.Println("host:", req.Host)
	fmt.Println("method:", req.Method)
	resp, err = c.rt.RoundTrip(req)
	duration := time.Since(startTime)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.Status)
	}
	fmt.Println("sum use:", duration)
	return resp, err
}
