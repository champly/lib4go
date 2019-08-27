package remote

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const expireTime = 60 * 60

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
	closeCh        chan struct{}
)

func init() {
	sshClientPool = map[string]*clientModel{}
	sftpClientPool = map[string]*sftpClientModel{}
	closeCh = make(chan struct{})

	go loopDeleteSSHClient(closeCh)
	go loopDeleteSftpClient(closeCh)
}

func getSession(info *ServerInfo) (*ssh.Session, error) {
	client, err := getSSHClient(info)
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	if err == nil {
		return session, nil
	}
	if !strings.EqualFold(err.Error(), "EOF") {
		return nil, err
	}

	fmt.Printf("远程主机(%s)已经关闭，重新开始连接\n", info.Host)
	closeClient(info.Host)

	client, err = getSSHClient(info)
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

	fmt.Println("构建新连接:", info.Host)
	auth := make([]ssh.AuthMethod, 0)
	if info.Key == "" {
		auth = append(auth, ssh.Password(info.Password))
	} else {
		var signer ssh.Signer
		var err error
		if info.Password == "" {
			signer, err = ssh.ParsePrivateKey([]byte(info.Key))
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(info.Key), []byte(info.Password))
		}
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	config := &ssh.ClientConfig{
		User: info.User,
		Auth: auth,
		Config: ssh.Config{
			Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		},
		Timeout: 10 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	c, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", info.Host, 22), config)
	if err != nil {
		return nil, err
	}
	l.Lock()
	sshClientPool[info.Host] = &clientModel{expireTime: time.Now().Add(time.Second * expireTime), client: c}
	l.Unlock()
	return c, nil
}

func getSftpClient(info *ServerInfo) (*sftp.Client, error) {
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

	l.Lock()
	sftpClientPool[info.Host] = &sftpClientModel{expireTime: time.Now().Add(time.Second * expireTime), client: sc}
	l.Unlock()
	return sc, nil
}

func loopDeleteSSHClient(closeCh chan struct{}) {
	ticker := time.NewTicker(time.Second * expireTime)
	for {
		select {
		case <-ticker.C:
			l.Lock()
			for host, model := range sshClientPool {
				if time.Now().Before(model.expireTime) {
					continue
				}

				// fmt.Println("自动回收ssh")
				model.client.Close()
				delete(sshClientPool, host)
			}
			l.Unlock()

		case <-closeCh:
			return
		}
	}
}

func loopDeleteSftpClient(closeCh chan struct{}) {
	ticker := time.NewTicker(time.Second * expireTime)
	for {
		select {
		case <-ticker.C:
			l.Lock()
			for host, model := range sftpClientPool {
				if time.Now().Before(model.expireTime) {
					continue
				}

				// fmt.Println("自动回收sftp")
				model.client.Close()
				delete(sftpClientPool, host)
			}
			l.Unlock()
		case <-closeCh:
			return
		}
	}
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
	close(closeCh)

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
