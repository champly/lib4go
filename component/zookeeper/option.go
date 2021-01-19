package zookeeper

import (
	"math"
	"time"
)

type Option func(*option)

type option struct {
	connectNum     int
	execTimeout    time.Duration
	connectTimeout time.Duration
}

func DefaultOption() *option {
	return &option{
		connectNum:     1,
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

// WithConnectNum set connect num
func WithConnectNum(num int) func(*option) {
	if num < 0 || num > math.MaxInt32 {
		num = 1
	}
	return func(opt *option) {
		opt.connectNum = num
	}
}
