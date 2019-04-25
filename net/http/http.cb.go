package http

import (
	"net/http"
)

type IHttpCb interface {
	Before(req *http.Request)
	After(resp *http.Response, err error)
}

type callChian struct {
	rt http.RoundTripper
	cb IHttpCb
}

func NewClientWithCb(cb IHttpCb) *http.Client {
	return &http.Client{
		Transport: newCbRoundTrip(http.DefaultTransport, cb),
	}
}

func newCbRoundTrip(rt http.RoundTripper, cb IHttpCb) http.RoundTripper {
	return &callChian{
		rt: rt,
		cb: cb,
	}
}

func (c *callChian) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	c.cb.Before(req)
	resp, err = c.rt.RoundTrip(req)
	c.cb.After(resp, err)
	return resp, err
}
