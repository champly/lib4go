package remote

import (
	"testing"
)

func TestExec(t *testing.T) {
	// create client
	client, err := NewRemoteClient(&ServerInfo{
		Host:     "10.13.3.3",
		User:     "root",
		Password: "dmallk8s",
		Port:     22,
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
	if err = client.ScpFile("./remote.go", "/root/tmp/src/remote.go"); err != nil {
		t.Error(err)
	}

	// scp dir
	if err = client.ScpDir("/Users/champly/Downloads/bak/rpm", "/root/tmp/rpm"); err != nil {
		t.Error(err)
	}

	// scp bash an exec
	r, err = client.UseBashExecScript("/root/tmp/exec.sh", "#!/bin/bash\ndate")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(r)
}
