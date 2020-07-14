package zookeeper

import "time"

type Option func(*option)

type option struct {
	execTimeout    time.Duration
	connectTimeout time.Duration
}

func DefaultOption() *option {
	return &option{
		execTimeout:    time.Second * 1,
		connectTimeout: time.Second * 5,
	}
}

// WithExecTimeout set exec timeout
func WithExecTimeout(timeout time.Duration) func(*option) {
	return func(opt *option) {
		opt.execTimeout = timeout
	}
}

// WithConnectTimeout set connect timeout
func WithConnectTimeout(timeout time.Duration) func(*option) {
	return func(opt *option) {
		opt.connectTimeout = timeout
	}
}
