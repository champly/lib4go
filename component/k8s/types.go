package k8s

import (
	"time"

	"k8s.io/client-go/rest"
)

const (
	// GracefulStopWaitTimeout graceful stop cluster wait time
	GracefulStopWaitTimeout = time.Second * 30
)

// Cluster status
const (
	Initing = iota
	Connected
	DisConnected
)

type KubeConfigType string

// Kubeconfig type
const (
	KubeConfigTypeRawString KubeConfigType = "RawString"
	KubeConfigTypeFile      KubeConfigType = "File"
)

// SetKubeRestConfigFn set kubernetes restconfig info
type SetKubeRestConfigFn func(*rest.Config)

// InitHandler ...
type InitHandler func(*Client) error

// ClusterConfigInfo cluster config info
type ClusterConfigInfo interface {
	GetName() string
	GetKubeConfig() string
	GetKubeContext() string
	GetKubeConfigType() KubeConfigType
}

// ClusterConfigurationManager cluster configuration manager
type ClusterConfigurationManager interface {
	GetAll() ([]ClusterConfigInfo, error)
	GetOptions() []Option
}
