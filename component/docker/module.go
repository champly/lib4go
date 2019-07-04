package docker

import (
	"fmt"
)

const DefaultRepository = "docker.io/library"

type ImageInfo struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageID    string `json:"image_id"`
	Created    string `json:"created"`
	Size       string `json:"size"`
}

type ContainerInfo struct {
	ID         string      `json:"id"`
	Names      []string    `json:"names"`
	Image      string      `json:"image"`
	ImageID    string      `json:"image_id"`
	Command    string      `json:"command"`
	SizeRw     string      `json:"size_rw,omitempty"`
	SizeRootFs string      `json:"size_rootfs,omitempty"`
	Created    string      `json:"created"`
	State      string      `json:"state"`
	Status     string      `json:"status"`
	Ports      []PortInfo  `json:"ports"`
	Mounts     []MountInfo `json:"mounts"`
}

type PortInfo struct {
	IP          string `json:"ip,omitempty"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port,omitempty"`
	Type        string `json:"type"`
}

type MountInfo struct {
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Driver      string `json:"driver,omitempty"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
	Propagation string `json:"propagation"`
}

const unit = 1000

func GetSize(byt int64) string {
	if byt < unit {
		return fmt.Sprintf("%dB", byt)
	}

	f := float64(byt) / unit
	if f < unit {
		return fmt.Sprintf("%.2fKB", f)
	}

	f /= unit
	if f < unit {
		return fmt.Sprintf("%.2fMB", f)
	}

	f /= unit
	if f < unit {
		return fmt.Sprintf("%.2fGB", f)
	}

	return fmt.Sprintf("%dB", byt)
}
