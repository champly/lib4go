package remote

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	clientList     map[string]*ssh.Client
	sftpClientList map[string]*sftp.Client
	l              sync.Mutex
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
}

func init() {
	clientList = map[string]*ssh.Client{}
	sftpClientList = map[string]*sftp.Client{}
}

func getSSHConnect(info *ServerInfo) (*ssh.Client, error) {
	if c, ok := clientList[info.Host]; ok {
		return c, nil
	}

	config := &ssh.ClientConfig{
		User: info.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(info.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	c, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", info.Host, info.Port), config)
	if err != nil {
		return nil, err
	}
	clientList[info.Host] = c
	return c, nil
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
	fmt.Println("cmd:", cmd)
	r.closeSftpClient()

	session, err := r.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := session.StderrPipe()
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
			break
		}
		bf.Write(buf[:n])
	}

	reader = bufio.NewReader(stderr)
	bfe := new(bytes.Buffer)
	bufe := make([]byte, 1024)
	for {
		n, err := reader.Read(bufe)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			if strings.TrimSpace(bfe.String()) == "" {
				return bf.String(), nil
			}
			if strings.TrimSpace(bf.String()) != "" {
				return bf.String() + bfe.String(), nil
			}
			return bf.String(), errors.New(bfe.String())
		}
		bfe.Write(bufe[:n])
	}
}

func (r *RemoteClient) getSftpClient() error {
	if r.sftpClient != nil {
		return nil
	}
	l.Lock()
	defer l.Unlock()

	if r.sftpClient != nil {
		return nil
	}

	if sc, ok := sftpClientList[r.Host]; ok {
		r.sftpClient = sc
		return nil
	}
	sc, err := sftp.NewClient(r.client)
	if err != nil {
		return err
	}
	sftpClientList[r.Host] = sc
	r.sftpClient = sc
	return nil
}

func (r *RemoteClient) closeSftpClient() {
	delete(sftpClientList, r.Host)
	if r.sftpClient != nil {
		r.sftpClient.Close()
		r.sftpClient = nil
	}
	return
}

func (r *RemoteClient) ScpFile(file string, remoteFile string) error {
	fmt.Println("scp file:", file, remoteFile)
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
	fmt.Println("scp dir:", localDir, remoteDir)
	if err := r.getSftpClient(); err != nil {
		return err
	}

	localDir = strings.TrimRight(localDir, "/")
	remoteDir = strings.TrimRight(remoteDir, "/")

	dir, err := ioutil.ReadDir(localDir)
	if err != nil {
		return err
	}

	r.sftpClient.MkdirAll(remoteDir)
	for _, f := range dir {
		rf := fmt.Sprintf(fmt.Sprintf("%s/%s", remoteDir, f.Name()))
		lf := fmt.Sprintf(fmt.Sprintf("%s/%s", localDir, f.Name()))

		if f.IsDir() {
			r.sftpClient.MkdirAll(rf)
			if err = r.ScpDir(lf, rf); err != nil {
				fmt.Println(err)
				return err
			}
			continue
		}
		r.ScpFile(lf, rf)
	}
	return nil
}

func (r *RemoteClient) CopyFile(remoteFile string, localFile string) error {
	fmt.Println("copy file:", remoteFile, localFile)
	if err := r.getSftpClient(); err != nil {
		return err
	}

	rf, err := r.sftpClient.OpenFile(remoteFile, os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("remote read file %s fail:%s", remoteFile, err)
	}

	b, err := ioutil.ReadAll(rf)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Base(localFile), os.ModePerm.Perm()); err != nil {
		return err
	}

	lf, err := os.OpenFile(localFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("create or read local file %s fail:%s", localFile, err)
	}
	defer lf.Close()

	lf.Write(b)
	return nil
}

func (r *RemoteClient) CopyDir(remoteDir, localDir string) error {
	fmt.Println("copy dir:", remoteDir, localDir)
	if err := r.getSftpClient(); err != nil {
		return err
	}

	dir, err := r.sftpClient.ReadDir(remoteDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(localDir, os.ModePerm.Perm()); err != nil {
		return err
	}
	for _, f := range dir {
		rf := fmt.Sprintf(fmt.Sprintf("%s/%s", remoteDir, f.Name()))
		lf := fmt.Sprintf(fmt.Sprintf("%s/%s", localDir, f.Name()))

		if f.IsDir() {
			if err := os.MkdirAll(lf, os.ModePerm.Perm()); err != nil {
				return err
			}

			if err = r.CopyDir(rf, lf); err != nil {
				fmt.Println(err)
				return err
			}
			continue
		}
		if err := r.CopyFile(rf, lf); err != nil {
			return err
		}
	}
	return nil
}

func (r *RemoteClient) UseBashExecScript(remoteFile, script string) (string, error) {
	fmt.Println("sh:", remoteFile)
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
	l.Lock()
	defer l.Unlock()

	r.closeSftpClient()
	delete(clientList, r.Host)
	r.client.Close()
}

func Close() {
	l.Lock()
	defer l.Unlock()

	for _, sc := range sftpClientList {
		sc.Close()
	}

	for _, c := range clientList {
		c.Close()
	}
}
