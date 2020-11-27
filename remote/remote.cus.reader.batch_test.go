package remote

import (
	"fmt"
	"testing"
	"time"
)

func TestBatchCusReader(t *testing.T) {
	client, err := NewBatchCusReader([]*ServerInfo{
		{
			Host: "10.12.192.131",
			User: "root",
			Key:  rsaPriv,
			Port: 22,
		},
		{
			Host: "10.12.192.132",
			User: "root",
			Key:  rsaPriv,
			Port: 22,
		},
		{
			Host: "10.12.192.133",
			User: "root",
			Key:  rsaPriv,
			Port: 22,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("connect success")

	go func() {
		for {
			rl := client.ExecGetResult()
			fmt.Println("go func:")
			for _, r := range rl {
				fmt.Println(r.Host, r.Msg)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// exec cmd
	// rl, err := client.Exec("for i in `seq 1 10`;do echo $i; sleep 1;done")
	// rl, err := client.Exec("docker pull nginx")
	rl, err := client.UseBashExecScript("/root/1.sh", `#!/bin/bash
echo "123"
sleep 1
echo "123"
sleep 1
echo "123"
sleep 1
echo "123"
sleep 1
echo "123"
sleep 1
	`)
	if err != nil {
		t.Error(err)
		return
	}

	for _, r := range rl {
		fmt.Println(r.Host, r.Msg)
	}
}
