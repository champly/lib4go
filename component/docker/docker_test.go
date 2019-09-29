package docker

import (
	"testing"
)

func TestGetAllImage(t *testing.T) {
	client, err := NewClient("10.12.192.130", 2376, "v1.39")
	if err != nil {
		t.Error(err)
	}
	list, err := client.GetAllImageList()
	if err != nil {
		t.Error(err)
	}

	for _, image := range list {
		t.Logf("%+v\n", image)
	}
}

func TestGetAllContainer(t *testing.T) {
	client, err := NewClient("10.12.192.130", 2376, "v1.39")
	if err != nil {
		t.Error(err)
	}
	list, err := client.GetAllContainer()
	if err != nil {
		t.Error(err)
	}

	for _, container := range list {
		t.Log(container)
	}
}

func TestPullImage(t *testing.T) {
	client, err := NewClient("10.12.192.130", 2376, "v1.39")
	if err != nil {
		t.Error(err)
	}
	t.Log(client.PullImage(DefaultRepository, "nginx", ""))
}

func TestCreateContainer(t *testing.T) {
	client, err := NewClient("10.12.192.130", 2376, "v1.39")
	if err != nil {
		t.Error(err)
	}
	ports := map[int][]int{
		8080:  []int{8080},
		50000: []int{50000},
	}
	id, err := client.CreateContainer(DefaultRepository, "jenkins", "latest", []string{"/root/jenkins:/root/123"}, ports, "jjjj")
	if err != nil {
		t.Error(err)
	}
	t.Log(id)

	t.Log(client.StartContainer(id))
}

// func TestTag(t *testing.T) {
// client, err := NewClient("10.12.192.130", 2376, "v1.39")
// if err != nil {
// t.Error(err)
// }
// t.Log(client.Tag("nginx:latest", "nginx:v1.23"))
// }
