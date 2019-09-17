package remote

import (
	"bytes"
	"fmt"

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
	session, err := getSession(info)
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
