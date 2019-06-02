package remote

import (
	"testing"
)

func TestBatchExec(t *testing.T) {
	// create client
	client, err := NewBatchRemoteClient([]*ServerInfo{
		&ServerInfo{
			Host:     "10.12.194.36",
			User:     "root",
			Password: "dmallk8s",
			Port:     22,
		},
		&ServerInfo{
			Host:     "10.12.194.94",
			User:     "root",
			Password: "dmallk8s",
			Port:     22,
		},
		&ServerInfo{
			Host:     "10.12.194.105",
			User:     "root",
			Password: "dmallk8s",
			Port:     22,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()
	t.Log("connect success")

	// exec cmd
	r, err := client.Exec("ls /")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(r)

	r, err = client.Exec("date")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(r)

	// scp file
	r, err = client.ScpFile("./remote.go", "/root/tmp/src/remote.go")
	if err != nil {
		t.Error(err)
	}
	t.Log(r)

	// scp dir
	r, err = client.ScpDir("/Users/champly/Downloads/bak/rpm", "/root/tmp/rpm")
	if err != nil {
		t.Error(err)
	}
	t.Log(r)

	// scp bash an exec
	r, err = client.UseBashExecScript("/root/tmp/exec.sh", "#!/bin/bash\ndate")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(r)
}
