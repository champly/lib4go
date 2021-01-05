package k8s

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type option struct {
	stopCh                  chan struct{}
	kubeConfig              string
	kubeContext             string
	kubeConfigType          KubeConfigType
	ctrlRtManagerOpts       manager.Options
	setKubeRestConfigFnList []SetKubeRestConfigFn
	healthCheckInterval     time.Duration
	requestTimeout          time.Duration
	resyncInterval          time.Duration

	ConnectStatus int
	StartStatus   bool
}

func buildDefaultCfg() *option {
	return &option{
		stopCh: make(chan struct{}, 0),
		ctrlRtManagerOpts: manager.Options{
			LeaderElection:         false,
			MetricsBindAddress:     "0",
			HealthProbeBindAddress: "0",
		},
		kubeConfigType:      KubeConfigTypeFile,
		healthCheckInterval: time.Second * 5,
		requestTimeout:      time.Second * 5,
		ConnectStatus:       Initing,
	}
}

type Option func(*option)

func WithKubeConfig(kubeConfig string) Option {
	return func(opt *option) {
		opt.kubeConfig = kubeConfig
	}
}

func WithKubeContext(kubeContext string) Option {
	return func(opt *option) {
		opt.kubeContext = kubeContext
	}
}

func WithKubeConfigType(typ KubeConfigType) Option {
	return func(opt *option) {
		opt.kubeConfigType = typ
	}
}

func WithKubeSetRsetConfigFn(setKubeRestConfigFnList ...SetKubeRestConfigFn) Option {
	return func(opt *option) {
		opt.setKubeRestConfigFnList = setKubeRestConfigFnList
	}
}

func WithRuntimeManagerOptions(ctrlRtManagerOpts manager.Options) Option {
	return func(opt *option) {
		opt.ctrlRtManagerOpts = ctrlRtManagerOpts
	}
}

func WithAutoHealthCheckInterval(interval time.Duration) Option {
	return func(opt *option) {
		opt.healthCheckInterval = interval
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(opt *option) {
		opt.requestTimeout = timeout
	}
}
