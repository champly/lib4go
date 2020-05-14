package remote

import (
	"testing"
)

func TestGetSSHClient(t *testing.T) {
	_, err := GetSSHClient(&ServerInfo{
		User:     "root",
		Password: "123456",
		Host:     "192.168.177.128",
	})
	if err != nil {
		t.Error(err)
	}
}
