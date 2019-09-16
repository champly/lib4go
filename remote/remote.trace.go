package remote

import (
	"bytes"
	"sync"

	"golang.org/x/crypto/ssh"
)

type TraceObj struct {
	session   *ssh.Session
	content   chan string
	contBuf   *bytes.Buffer
	l         sync.Mutex
	needClose bool
}

func NewRemoteTrace(info *ServerInfo) (*TraceObj, error) {
	session, err := getSession(info)
	if err != nil {
		return nil, err
	}

	return &TraceObj{
		session: session,
		content: make(chan string, 2),
		contBuf: new(bytes.Buffer),
	}, nil
}

func (r *TraceObj) Exec(cmd string) (string, error) {
	r.session.Stdout = r
	r.session.Stderr = r

	r.session.Run(cmd)
	return r.contBuf.String(), nil
}

func (r *TraceObj) ExecGetResult() (string, error) {
	return r.contBuf.String(), nil
}

func (r *TraceObj) Write(p []byte) (n int, err error) {
	r.contBuf.Write(p)
	return len(p), nil
}
