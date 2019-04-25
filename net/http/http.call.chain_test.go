package http

import "testing"

func TestNewClientWithCallChain(t *testing.T) {
	client := NewClientWithCallChain()
	client.Get("http://www.baidu.com")
}
