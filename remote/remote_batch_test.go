package remote

// func TestBatchExec(t *testing.T) {
// // create client
// client, err := NewBatchRemoteClient([]*ServerInfo{
// &ServerInfo{
// Host:     "10.13.3.6",
// User:     "root",
// Password: "dmallk8s",
// Port:     22,
// },
// // &ServerInfo{
// // Host:     "10.13.3.4",
// // User:     "root",
// // Password: "dmallk8s",
// // Port:     22,
// // },
// // &ServerInfo{
// // Host:     "10.13.3.5",
// // User:     "root",
// // Password: "dmallk8s",
// // Port:     22,
// // },
// })
// if err != nil {
// t.Error(err)
// return
// }
// defer client.Close()
// t.Log("connect success")

// // exec cmd
// r, err := client.Exec("ls /")
// if err != nil {
// t.Error(err)
// return
// }
// t.Log(r)

// r, err = client.Exec("date")
// if err != nil {
// t.Error(err)
// return
// }
// t.Log(r)

// // scp file
// r, err = client.ScpFile("./remote.go", "/root/tmp/src/remote.go")
// if err != nil {
// t.Error(err)
// return
// }
// t.Log(r)

// // scp dir
// r, err = client.ScpDir("/Users/champly/Documents/kops/test/k8s", "/root/tmp/rpm")
// if err != nil {
// t.Error(err)
// return
// }
// t.Log(r)

// // scp bash an exec
// r, err = client.UseBashExecScript("/root/tmp/exec.sh", "#!/bin/bash\ndate")
// if err != nil {
// t.Error(err)
// return
// }
// t.Log(r)
// }
