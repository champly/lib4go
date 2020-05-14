package remote

import (
	"testing"
	"time"
)

func TestGetSSHClient(t *testing.T) {
	_, err := GetSSHClient(&ServerInfo{
		User:     "root",
		Password: "123456",
		Host:     "192.168.177.128",
		Port:     22,
	})
	if err != nil {
		t.Error(err)
	}

	time.Sleep(60 * time.Second)
}
