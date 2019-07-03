package docker

import "testing"

func TestNewClient(t *testing.T) {
	client, err := NewClient("10.13.3.2", 2376)
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
