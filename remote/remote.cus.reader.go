package remote

import (
	"bytes"
	"fmt"
	"path"

	"golang.org/x/crypto/ssh"
)

type CusReader struct {
	*ServerInfo
	session *ssh.Session
	contBuf *bytes.Buffer
}

func NewCusReaderWithSession(info *ServerInfo, session *ssh.Session) *CusReader {
	return &CusReader{
		ServerInfo: info,
		session:    session,
		contBuf:    new(bytes.Buffer),
	}
}

func (c *CusReader) Write(p []byte) (n int, err error) {
	c.contBuf.Write(p)
	return len(p), nil
}

func (c *CusReader) Close() error {
	return c.session.Close()
}

func NewCusReader(info *ServerInfo) (*CusReader, error) {
	session, err := GetSession(info)
	if err != nil {
		return nil, err
	}

	return NewCusReaderWithSession(info, session), nil
}

func (c *CusReader) Exec(cmd string) (string, error) {
	c.session.Stdout = c
	c.session.Stderr = c

	err := c.session.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("%s%s", c.contBuf.String(), err.Error())
	}
	return c.contBuf.String(), nil
}

func (c *CusReader) ExecGetResult() string {
	return c.contBuf.String()
}

func (c *CusReader) UseBashExecScript(remoteFile, script string) (string, error) {
	sclient, err := GetSftpClient(c.ServerInfo)
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
	return c.Exec("sh " + remoteFile)
}
