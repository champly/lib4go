package remote

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const expireTime = 60

type clientModel struct {
	expireTime time.Time
	client     *ssh.Client
}

type sftpClientModel struct {
	expireTime time.Time
	client     *sftp.Client
}

var (
	sshClientPool  map[string]*clientModel
	sftpClientPool map[string]*sftpClientModel
	l              sync.Mutex
)

func init() {
	sshClientPool = map[string]*clientModel{}
	sftpClientPool = map[string]*sftpClientModel{}

	go loopDeleteSSHClient()
	go loopDeleteSftpClient()
}

func getSession(info *ServerInfo) (*ssh.Session, error) {
	client, err := getSSHClient(info)
	if err != nil {
		return nil, err
	}
	return client.NewSession()
}

func getSSHClient(info *ServerInfo) (*ssh.Client, error) {
	if c, ok := sshClientPool[info.Host]; ok {
		c.expireTime = time.Now().Add(time.Second * expireTime)
		return c.client, nil
	}
	fmt.Println("构建新连接")
	fmt.Println(info)

	config := &ssh.ClientConfig{
		User: info.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(info.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	c, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", info.Host, 22), config)
	if err != nil {
		return nil, err
	}
	sshClientPool[info.Host] = &clientModel{expireTime: time.Now().Add(time.Second * expireTime), client: c}
	return c, nil
}

func getSftpClient(info *ServerInfo) (*sftp.Client, error) {
	l.Lock()
	defer l.Unlock()

	if sc, ok := sftpClientPool[info.Host]; ok {
		sc.expireTime = time.Now().Add(time.Second * expireTime)
		return sc.client, nil
	}

	client, err := getSSHClient(info)
	if err != nil {
		return nil, err
	}
	sc, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}
	sftpClientPool[info.Host] = &sftpClientModel{expireTime: time.Now().Add(time.Second * expireTime), client: sc}
	return sc, nil
}

func loopDeleteSSHClient() {
	l.Lock()
	for host, model := range sshClientPool {
		if time.Now().Before(model.expireTime) {
			continue
		}

		fmt.Println("自动回收ssh")
		model.client.Close()
		delete(sshClientPool, host)
	}
	l.Unlock()

	time.Sleep(time.Second * expireTime)
}

func loopDeleteSftpClient() {
	l.Lock()
	for host, model := range sftpClientPool {
		if time.Now().Before(model.expireTime) {
			continue
		}

		fmt.Println("自动回收sftp")
		model.client.Close()
		delete(sftpClientPool, host)
	}
	l.Unlock()

	time.Sleep(time.Second * expireTime)
}

func closeClient(host string) {
	l.Lock()
	defer l.Unlock()

	client, ok := sshClientPool[host]
	if !ok {
		return
	}
	client.client.Close()
	delete(sshClientPool, host)
}

func closeSftpClient(host string) {
	l.Lock()
	defer l.Unlock()

	sftpModel, ok := sftpClientPool[host]
	if !ok {
		return
	}
	sftpModel.client.Close()
	delete(sftpClientPool, host)
}

func Close() {
	l.Lock()
	defer l.Unlock()

	for _, c := range sshClientPool {
		c.client.Close()
	}
	sshClientPool = nil

	for _, sc := range sftpClientPool {
		sc.client.Close()
	}
	sftpClientPool = nil
}
