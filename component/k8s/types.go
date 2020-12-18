package k8s

import "k8s.io/client-go/rest"

type ConnectStatus string

const (
	Initing      ConnectStatus = "initing"
	Connected    ConnectStatus = "connected"
	DisConnected ConnectStatus = "disconnected"
)

type RestConfigFunc func(*rest.Config)
