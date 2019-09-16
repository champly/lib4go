package remote

import (
	"fmt"
	"testing"
	"time"
)

func TestTraceObj(t *testing.T) {
	client, err := NewRemoteTrace(&ServerInfo{
		Host:     "192.168.50.92",
		User:     "root",
		Password: "123456",
		Port:     22,
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("connect success")

	go func() {
		for {
			r, _ := client.ExecGetResult()
			fmt.Println("go func:", r)
			time.Sleep(1 * time.Second)
		}
	}()

	// exec cmd
	// r, err := client.Exec("for i in `seq 1 10`;do echo $i; sleep 1;done")
	r, err := client.Exec("docker pull nginx")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(r)
}
