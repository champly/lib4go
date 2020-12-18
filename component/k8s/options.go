package k8s

import "time"

type option struct {
	kubeconfig string
	fns        []RestConfigFunc
	resync     time.Duration
}

func buildDefaultOpt() *option {
	return &option{
		resync: 0,
	}
}

type Option func(*option)

func WithKubeConfig(kubeconfig string) Option {
	return func(opt *option) {
		opt.kubeconfig = kubeconfig
	}
}

func WithResetConfigFunc(fns []RestConfigFunc) Option {
	return func(opt *option) {
		opt.fns = fns
	}
}
