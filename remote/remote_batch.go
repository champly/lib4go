package remote

import (
	"fmt"
	"sync"
)

type ResponseMsg struct {
	Host  string
	Msg   string
	Error error
}

type BatchRemoteClient struct {
	client []*RemoteClient
	count  int
	wg     sync.WaitGroup
	l      sync.Mutex
}

func NewBatchRemoteClient(serverList []*ServerInfo) (*BatchRemoteClient, error) {
	if serverList == nil || len(serverList) < 1 {
		return nil, fmt.Errorf("serverList must not nil")
	}
	rclient := &BatchRemoteClient{count: len(serverList)}

	var err error
	for _, serverInfo := range serverList {
		rclient.wg.Add(1)
		go func(serverInfo *ServerInfo) {
			defer rclient.wg.Done()
			var c *RemoteClient
			c, err = NewRemoteClient(serverInfo)
			if err != nil {
				return
			}
			rclient.client = append(rclient.client, c)

		}(serverInfo)
	}
	rclient.wg.Wait()
	if err != nil {
		return nil, err
	}

	return rclient, nil
}

func (b *BatchRemoteClient) Exec(cmd string) ([]ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]ResponseMsg, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			r, e := c.Exec(cmd)
			rsl = append(rsl, ResponseMsg{Msg: r, Error: e, Host: c.Host})
			b.wg.Done()
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) ScpFile(file string, remoteFile string) ([]ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]ResponseMsg, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			e := c.ScpFile(file, remoteFile)
			rsl = append(rsl, ResponseMsg{Error: e, Host: c.Host})
			b.wg.Done()
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) ScpDir(localDir, remoteDir string) ([]ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]ResponseMsg, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			e := c.ScpDir(localDir, remoteDir)
			rsl = append(rsl, ResponseMsg{Error: e, Host: c.Host})
			b.wg.Done()
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchRemoteClient) UseBashExecScript(remoteFile, script string) ([]ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]ResponseMsg, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *RemoteClient) {
			r, e := c.UseBashExecScript(remoteFile, script)
			rsl = append(rsl, ResponseMsg{Msg: r, Error: e, Host: c.Host})
			b.wg.Done()
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
