package remote

import (
	"fmt"
	"sync"
)

type BatchCusReader struct {
	client []*CusReader
	count  int
	l      sync.Mutex
	wg     sync.WaitGroup
}

func NewBatchCusReader(serverList []*ServerInfo) (*BatchCusReader, error) {
	if serverList == nil || len(serverList) < 1 {
		return nil, fmt.Errorf("serverList must not nil")
	}
	rclient := &BatchCusReader{count: len(serverList)}

	errlist := make([]error, len(serverList))
	rclient.client = []*CusReader{}
	for i, serverInfo := range serverList {
		rclient.wg.Add(1)
		go func(serverInfo *ServerInfo, index int) {
			defer rclient.wg.Done()

			c, err := NewCusReader(serverInfo)
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

func (b *BatchCusReader) Exec(cmd string) ([]*ResponseMsg, error) {
	b.l.Lock()
	defer b.l.Unlock()

	rsl := make([]*ResponseMsg, 0, b.count)

	b.wg.Add(b.count)
	for _, c := range b.client {
		go func(c *CusReader) {
			defer b.wg.Done()
			r, e := c.Exec(cmd)
			rsl = append(rsl, &ResponseMsg{Msg: r, Error: e, Host: c.Host})
		}(c)
	}
	b.wg.Wait()
	return rsl, nil
}

func (b *BatchCusReader) ExecGetResult() []*ResponseMsg {
	rsl := make([]*ResponseMsg, 0, b.count)

	for _, c := range b.client {
		rsl = append(rsl, &ResponseMsg{Host: c.Host, Msg: c.ExecGetResult()})
	}
	return rsl
}
