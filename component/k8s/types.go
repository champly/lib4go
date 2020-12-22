package k8s

import (
	"time"

	"k8s.io/client-go/rest"
)

const (
	GracefulStopWaitTimeout = time.Second * 30
)

const (
	Initing = iota
	Connected
	DisConnected
)

type RestConfigFunc func(*rest.Config)
type BeforeStartFunc func(*Client) error

type ClusterInfo interface {
	GetName() string
	GetKubeConfig() string
	GetOptions() []Option
}

type ClusterConfiguration interface {
	GetAll() []ClusterInfo
}
