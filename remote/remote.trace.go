package remote

import (
	"bytes"

	"golang.org/x/crypto/ssh"
)

type TraceObj struct {
	session *ssh.Session
	contBuf *bytes.Buffer
}

func NewRemoteTrace(info *ServerInfo) (*TraceObj, error) {
	session, err := getSession(info)
	if err != nil {
		return nil, err
	}

	return &TraceObj{
		session: session,
		contBuf: new(bytes.Buffer),
	}, nil
}

func NewRemoteTraceWithSession(session *ssh.Session) *TraceObj {
	return &TraceObj{
		session: session,
		contBuf: new(bytes.Buffer),
	}
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

func (r *TraceObj) Close() error {
	return r.session.Close()
}
