package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type DClient struct {
	cli *client.Client
}

func NewClient(host string, port int) (*DClient, error) {
	client, err := client.NewClient(fmt.Sprintf("tcp://%s:%d", host, port), "", nil, nil)
	if err != nil {
		return nil, err
	}
	return &DClient{cli: client}, nil
}

func (d *DClient) CreateContainer(images string, volument []string, port map[int][]int, cname string) (id string, err error) {
	return
}

func (d *DClient) GetAllImageList() ([]ImageInfo, error) {
	list, err := d.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, fmt.Errorf("[GetAllImageList] failed:%s", err.Error())
	}

	imageList := []ImageInfo{}
	for _, image := range list {
		imageList = append(imageList, ImageInfo{image})
	}
	return imageList, nil
}

func (d *DClient) GetAllContainer() ([]ContainerInfo, error) {
	list, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, fmt.Errorf("[GetAllContainer] failed:%s", err.Error())
	}

	containerList := []ContainerInfo{}
	for _, container := range list {
		containerList = append(containerList, ContainerInfo{container})
	}
	return containerList, nil
}

func (d *DClient) PullImage(imageName string) (isDownload bool, err error) {
	resp, err := d.cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
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

func (d *DClient) ImageExists(imageName string) (err error) {
	return
}

func (d *DClient) DeleteImage(imageName string) (err error) {
	return
}
