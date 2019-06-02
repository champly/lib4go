package remote

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"path"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ServerInfo struct {
	User     string
	Password string
	Host     string
	Port     int
}

type RemoteClient struct {
	*ServerInfo
	client     *ssh.Client
	sftpClient *sftp.Client
	l          sync.Mutex
}

func getSSHConnect(info *ServerInfo) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: info.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(info.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", info.Host, info.Port), config)
}

func NewRemoteClient(info *ServerInfo) (*RemoteClient, error) {
	client, err := getSSHConnect(info)
	if err != nil {
		return nil, err
	}
	rclient := &RemoteClient{
		client:     client,
		ServerInfo: info,
	}
	return rclient, nil
}

func (r *RemoteClient) Exec(cmd string) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", err
	}

	session.Run(cmd)
	reader := bufio.NewReader(stdout)
	bf := new(bytes.Buffer)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			return bf.String(), nil
		}
		bf.Write(buf[:n])
	}
}

func (r *RemoteClient) getSftpClient() error {
	if r.sftpClient != nil {
		return nil
	}

	r.l.Lock()
	defer r.l.Unlock()

	if r.sftpClient != nil {
		return nil
	}

	sc, err := sftp.NewClient(r.client)
	if err != nil {
		return err
	}
	r.sftpClient = sc
	return nil
}

func (r *RemoteClient) ScpFile(file string, remoteFile string) error {
	if err := r.getSftpClient(); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err = r.sftpClient.MkdirAll(path.Dir(remoteFile)); err != nil {
		return fmt.Errorf("create remote dir(%s) fail:%s", path.Dir(remoteFile), err)
	}

	dsf, err := r.sftpClient.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("create remote file fail:%s", err)
	}
	dsf.Write(b)
	return nil
}

func (r *RemoteClient) ScpDir(localDir, remoteDir string) error {
	if err := r.getSftpClient(); err != nil {
		return err
	}
	localDir = strings.TrimRight(localDir, "/")
	remoteDir = strings.TrimRight(remoteDir, "/")

	dir, err := ioutil.ReadDir(localDir)
	if err != nil {
		return err
	}
	for _, f := range dir {
		rf := fmt.Sprintf(fmt.Sprintf("%s/%s", remoteDir, f.Name()))
		lf := fmt.Sprintf(fmt.Sprintf("%s/%s", localDir, f.Name()))

		if f.IsDir() {
			r.sftpClient.MkdirAll(rf)
			if err = r.ScpDir(lf, rf); err != nil {
				return err
			}
			continue
		}
		r.sftpClient.MkdirAll(remoteDir)
		b, err := ioutil.ReadFile(lf)
		if err != nil {
			return err
		}
		dsf, err := r.sftpClient.Create(rf)
		if err != nil {
			return err
		}
		dsf.Write(b)
	}
	return nil
}

func (r *RemoteClient) CopyFile(file string, remoteFile string) error {
	return nil
}

func (r *RemoteClient) CopyDir(localDir, remoteDir string) error {
	return nil
}

func (r *RemoteClient) UseBashExecScript(remoteFile, script string) (string, error) {
	if err := r.getSftpClient(); err != nil {
		return "", err
	}

	if err := r.sftpClient.MkdirAll(path.Dir(remoteFile)); err != nil {
		return "", fmt.Errorf("create remote dir(%s) fail:%s", path.Dir(remoteFile), err)
	}

	dsf, err := r.sftpClient.Create(remoteFile)
	if err != nil {
		return "", fmt.Errorf("create remote file fail:%s", err)
	}
	dsf.Write([]byte(script))
	return r.Exec("sh " + remoteFile)
}

func (r *RemoteClient) Close() {
	r.client.Close()
}
