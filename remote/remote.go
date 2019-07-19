package remote

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type ServerInfo struct {
	User     string
	Password string
	Host     string
	Port     int
}

type RemoteClient struct {
	*ServerInfo
}

func NewRemoteClient(info *ServerInfo) (*RemoteClient, error) {
	_, err := getSSHClient(info)
	if err != nil {
		return nil, err
	}
	rclient := &RemoteClient{
		ServerInfo: info,
	}
	return rclient, nil
}

func (r *RemoteClient) Exec(cmd string) (string, error) {
	closeSftpClient(r.Host)

	session, err := getSession(r.ServerInfo)
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
			return "", errors.New(bfe.String())
		}
		bfe.Write(bufe[:n])
	}
}

func (r *RemoteClient) ExecWithTimeout(cmd string, t time.Duration) (string, error) {
	var result string
	var err error

	done := make(chan bool)
	go func() {
		result, err = r.Exec(cmd)
		done <- true
	}()

	select {
	case <-time.After(t):
		r.Close()
		return "", fmt.Errorf("exec timeout")
	case <-done:
		return result, err
	}
}

func (r *RemoteClient) ScpFile(file string, remoteFile string) error {
	sclient, err := getSftpClient(r.ServerInfo)

	if err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("read file:%s fail:%s", file, err)
	}
	defer f.Close()

	if err = sclient.MkdirAll(path.Dir(remoteFile)); err != nil {
		return fmt.Errorf("create remote dir(%s) fail:%s", path.Dir(remoteFile), err)
	}

	dsf, err := sclient.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("create remote file fail:%s", err)
	}
	defer dsf.Close()

	io.Copy(dsf, f)
	return nil
}

func (r *RemoteClient) ScpDir(localDir, remoteDir string) error {
	sclient, err := getSftpClient(r.ServerInfo)

	if err != nil {
		return err
	}

	localDir = strings.TrimRight(localDir, "/")
	remoteDir = strings.TrimRight(remoteDir, "/")

	dir, err := ioutil.ReadDir(localDir)
	if err != nil {
		return err
	}

	sclient.MkdirAll(remoteDir)
	for _, f := range dir {
		rf := fmt.Sprintf(fmt.Sprintf("%s/%s", remoteDir, f.Name()))
		lf := fmt.Sprintf(fmt.Sprintf("%s/%s", localDir, f.Name()))

		if f.IsDir() {
			sclient.MkdirAll(rf)
			if err = r.ScpDir(lf, rf); err != nil {
				return err
			}
			continue
		}
		r.ScpFile(lf, rf)
	}
	return nil
}

func (r *RemoteClient) CopyFile(localFile string, remoteFile string) error {
	sclient, err := getSftpClient(r.ServerInfo)
	if err != nil {
		return err
	}

	rf, err := sclient.OpenFile(remoteFile, os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("remote read file %s fail:%s", remoteFile, err)
	}

	b, err := ioutil.ReadAll(rf)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(localFile), os.ModePerm.Perm()); err != nil {
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

func (r *RemoteClient) CopyDir(localDir, remoteDir string) error {
	sclient, err := getSftpClient(r.ServerInfo)
	if err != nil {
		return err
	}

	dir, err := sclient.ReadDir(remoteDir)
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

			if err = r.CopyDir(lf, rf); err != nil {
				return err
			}
			continue
		}
		if err := r.CopyFile(lf, rf); err != nil {
			return err
		}
	}
	return nil
}

func (r *RemoteClient) UseBashExecScript(remoteFile, script string) (string, error) {
	sclient, err := getSftpClient(r.ServerInfo)
	if err != nil {
		return "", err
	}

	if err := sclient.MkdirAll(path.Dir(remoteFile)); err != nil {
		return "", fmt.Errorf("create remote dir(%s) fail:%s", path.Dir(remoteFile), err)
	}

	dsf, err := sclient.Create(remoteFile)
	if err != nil {
		return "", fmt.Errorf("create remote file fail:%s", err)
	}
	dsf.Write([]byte(script))
	return r.Exec("sh " + remoteFile)
}

func (r *RemoteClient) Close() {
	closeSftpClient(r.Host)
	closeClient(r.Host)
}
