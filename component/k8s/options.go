package k8s

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type option struct {
	gracefulStopCh chan struct{}
	clsname        string
	kubeconfig     string
	rtManagerOpts  manager.Options
	rsFns          []RestConfigFunc
	resync         time.Duration
	ConnectStatus  int
	StartStatus    bool
}

func getDefaultCfg() *option {
	return &option{
		gracefulStopCh: make(chan struct{}, 0),
		resync:         0,
		rtManagerOpts:  manager.Options{},
		ConnectStatus:  Initing,
	}
}

type Option func(*option)

func WithKubeConfig(kubeconfig string) Option {
	return func(opt *option) {
		opt.kubeconfig = kubeconfig
	}
}

func WithResetConfigFunc(rsFns ...RestConfigFunc) Option {
	return func(opt *option) {
		opt.rsFns = rsFns
	}
}

func WithRuntimeManagerOptions(rtManagerOpts manager.Options) Option {
	return func(opt *option) {
		opt.rtManagerOpts = rtManagerOpts
	}
}

func WithClusterName(name string) Option {
	return func(opt *option) {
		opt.clsname = name
	}
}
