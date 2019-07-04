package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DClient struct {
	cli *client.Client
}

func NewClient(host string, port int, version string) (*DClient, error) {
	client, err := client.NewClient(fmt.Sprintf("tcp://%s:%d", host, port), version, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DClient{cli: client}, nil
}

func (d *DClient) GetAllImageList() ([]ImageInfo, error) {
	list, err := d.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, fmt.Errorf("[GetAllImageList] failed:%s", err.Error())
	}

	imageList := []ImageInfo{}
	for _, image := range list {
		img := ImageInfo{
			ImageID: strings.Trim(image.ID, "sha256:"),
			Created: time.Unix(image.Created, 0).Format("2006-01-02 15:03:04"),
			Size:    GetSize(image.Size),
		}
		retags := image.RepoTags
		for _, retag := range retags {
			r := strings.Split(retag, ":")
			if len(r) != 2 {
				continue
			}
			img.Repository = r[0]
			img.Tag = r[1]
			imageList = append(imageList, img)
		}
	}
	return imageList, nil
}

func (d *DClient) PullImage(repository, imageName, version string) (isDownload bool, err error) {
	if version != "" {
		version = ":" + version
	}

	resp, err := d.cli.ImagePull(context.Background(), fmt.Sprintf("%s/%s%s", repository, imageName, version), types.ImagePullOptions{})
	if err != nil {
		return false, fmt.Errorf("[PullImage] image-name:%s failed:%s", imageName, err.Error())
	}

	buf := make([]byte, 1024)
	bs := bytes.Buffer{}
	for {
		n, err := resp.Read(buf)
		if err != nil && err != io.EOF {
			return false, fmt.Errorf("[PullImage] image-name:%s failed:%s", imageName, err.Error())
		}
		if n == 0 {
			break
		}
		bs.Write(buf[:n])
	}

	return strings.Contains(bs.String(), "Status: Downloaded newer image for"), nil
}

func (d *DClient) ImageIsExists(image, version string) (bool, error) {
	list, err := d.GetAllImageList()
	if err != nil {
		return false, err
	}
	for _, image := range list {
		if strings.EqualFold(fmt.Sprintf("%s:%s", image.Repository, image.Tag), fmt.Sprintf("%s:%s", image, version)) {
			return true, nil
		}
	}
	return false, nil
}

// func (d *DClient) Tag(sour, targ string) (err error) {
// return d.cli.ImageTag(context.Background(), sour, targ)
// }

func (d *DClient) GetAllContainer() ([]ContainerInfo, error) {
	list, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{
		Size: true,
		All:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("[GetAllContainer] failed:%s", err.Error())
	}

	containerList := []ContainerInfo{}
	for _, container := range list {
		cont := ContainerInfo{
			ID:         container.ID,
			Names:      container.Names,
			Image:      container.Image,
			ImageID:    strings.Trim(container.ID, "sha256:"),
			SizeRw:     GetSize(container.SizeRw),
			SizeRootFs: GetSize(container.SizeRootFs),
			Command:    container.Command,
			Created:    time.Unix(container.Created, 0).Format("2006-01-02 15:03:04"),
			State:      container.State,
			Status:     container.Status,
		}

		for _, port := range container.Ports {
			if cont.Ports == nil {
				cont.Ports = []PortInfo{}
			}
			cont.Ports = append(cont.Ports, PortInfo{
				IP:          port.IP,
				PrivatePort: port.PrivatePort,
				PublicPort:  port.PublicPort,
				Type:        port.Type,
			})
		}
		for _, mount := range container.Mounts {
			if cont.Mounts == nil {
				cont.Mounts = []MountInfo{}
			}
			cont.Mounts = append(cont.Mounts, MountInfo{
				Type:        string(mount.Type),
				Source:      mount.Source,
				Destination: mount.Destination,
				Mode:        mount.Mode,
				RW:          mount.RW,
				Propagation: string(mount.Propagation),
			})
		}
		containerList = append(containerList, cont)
	}
	return containerList, nil
}

func (d *DClient) CreateContainer(repository, imageName, version string, volument []string, ports map[int][]int, cname string) (id string, err error) {
	_, err = d.PullImage(repository, imageName, version)
	if err != nil {
		return "", err
	}

	if repository != DefaultRepository {
		imageName = repository + "/" + imageName
	}

	config := &container.Config{
		OpenStdin: true,
		Image:     fmt.Sprintf("%s:%s", imageName, version),
	}

	pbs := map[nat.Port][]nat.PortBinding{}
	for po, piList := range ports {

		pos := nat.Port(fmt.Sprintf("%d/tcp", po))
		pbs[pos] = []nat.PortBinding{}

		for _, pi := range piList {
			pbs[pos] = append(pbs[pos], nat.PortBinding{HostPort: fmt.Sprintf("%d", pi)})
		}
	}
	hostConfig := &container.HostConfig{
		Binds:        volument,
		PortBindings: pbs,
	}

	netConfig := &network.NetworkingConfig{}
	resp, err := d.cli.ContainerCreate(context.Background(), config, hostConfig, netConfig, cname)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *DClient) StartContainer(id string) error {
	return d.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
}
