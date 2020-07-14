package zookeeper

import "errors"

var (
	ErrClientDisConnect = errors.New("zk client is not connected to server")
	ErrExecTimeout      = errors.New("exec timeout")
)

const (
	Continue = true
	Stop     = false
)
