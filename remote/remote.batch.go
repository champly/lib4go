package remote

import (
	"fmt"
	"sync"
)

type ResponseMsg struct {
	Host  string `json:"host,omitempty"`
	Msg   string `json:"msg,omitempty"`
	Error error
}

type BatchRemoteClient struct {
	client []*RemoteClient
	count  int
	l      sync.Mutex
	wg     sync.WaitGroup
}

func NewBatchRemoteClient(serverList []*ServerInfo) (*BatchRemoteClient, error) {
	if serverList == nil || len(serverList) < 1 {
		return nil, fmt.Errorf("serverList must not nil")
	}
	rclient := &BatchRemoteClient{count: len(serverList)}

	errlist := make([]error, len(serverList))
	rclient.client = []*RemoteClient{}
	for i, serverInfo := range serverList {
		rclient.wg.Add(1)
		go func(serverInfo *ServerInfo, index int) {
			defer rclient.wg.Done()

			c, err := NewRemoteClient(serverInfo)
			if err != nil {
				errlist[index] = err
				return
			}
			rclient.l.Lock()
			rclient.client = append(rclient.client, c)
			rclient.l.Unlock()
			return
		}(serverInfo, i)
	}
	rclient.wg.Wait()

	for _, e := range errlist {
		if e != nil {
			return nil, e
		}
	}
	return rclient, nil
}

func (b *BatchRemoteClient) Exec(cmd string) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			defer b.wg.Done()
			r, e := c.Exec(cmd)
			rsl = append(rsl, &ResponseMsg{Msg: r, Error: e, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) ScpFile(localFile string, remoteFile string) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			defer b.wg.Done()
			e := c.ScpFile(localFile, remoteFile)
			rsl = append(rsl, &ResponseMsg{Error: e, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) ScpDir(localDir, remoteDir string) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			defer b.wg.Done()
			e := c.ScpDir(localDir, remoteDir)
			rsl = append(rsl, &ResponseMsg{Error: e, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) CopyFile(localFile string, remoteFile string) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			defer b.wg.Done()
			e := c.CopyFile(localFile, remoteFile)
			rsl = append(rsl, &ResponseMsg{Error: e, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) CopyDir(localDir, remoteDir string) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			defer b.wg.Done()
			e := c.CopyDir(localDir, remoteDir)
			rsl = append(rsl, &ResponseMsg{Error: e, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) UseBashExecScript(remoteFile, script string) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			defer b.wg.Done()
			r, e := c.UseBashExecScript(remoteFile, script)
			rsl = append(rsl, &ResponseMsg{Msg: r, Error: e, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) Foreach(f func(r *RemoteClient) (string, error)) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			defer b.wg.Done()
			str, err := f(c)
			rsl = append(rsl, &ResponseMsg{Msg: str, Error: err, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) Close() {
	b.l.Lock()
	defer b.l.Unlock()

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			c.Close()
			b.wg.Done()
		}(c)
	}
	b.wg.Wait()
	return
}
