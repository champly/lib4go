package remote

import (
	"fmt"
	"testing"
	"time"
)

func TestExec(t *testing.T) {
	// create client
	client, err := NewRemoteClient(&ServerInfo{
		Host: "10.12.192.131",
		User: "root",
		Key:  rsaPriv,
		Port: 22,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()
	fmt.Println("connect success")

	// exec cmd
	r, err := client.Exec("cd /root/demo && docker build -t demo .")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r)

	fmt.Println("等待关闭系统")
	time.Sleep(time.Second * 50)
	fmt.Println("开始执行下一步")

	r, err = client.Exec("cat /etc/passwd")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(r)

	// r, err = client.Exec("date")
	// if err != nil {
	// t.Error(err)
	// return
	// }
	// t.Log(r)

	// // scp file
	// if err = client.ScpFile("./remote.go", "/root/tmp/src/remote.go"); err != nil {
	// t.Error(err)
	// }

	// // scp dir
	// if err = client.ScpDir("/Users/champly/Downloads/bak/rpm", "/root/tmp/rpm"); err != nil {
	// t.Error(err)
	// }

	// // scp bash an exec
	// r, err = client.UseBashExecScript("/root/tmp/exec.sh", "#!/bin/bash\ndate")
	// if err != nil {
	// t.Error(err)
	// return
	// }

	// // scp dir
	// if err = client.ScpDir("/Users/champly/Documents/kops/test/k8s", "/root/tmp/rpm"); err != nil {
	// t.Error(err)
	// }

	// copy dir
	// if err = client.CopyDir("/etc/kubernetes", "/Users/champly/Downloads/kubernetes"); err != nil {
	// t.Error(err)
	// }
}
