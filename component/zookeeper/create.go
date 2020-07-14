package zookeeper

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
)

// CreatePersistentNode create persistent node with data
func (z *ZkClient) CreatePersistentNode(path string, data string) (err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	if strings.EqualFold(path, "/") {
		return nil
	}

	if !strings.HasPrefix(path, "/") {
		err = fmt.Errorf("path should start with '/':%s", path)
		return
	}

	b, e := z.IsExists(path)
	if e != nil {
		err = errors.Wrap(e, "create persistent node")
		return
	}
	if b {
		return nil
	}

	if err = z.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return err
	}

	e = WarpperTimeout(func() {
		_, err = z.conn.Create(path, []byte(data), int32(0), zk.WorldACL(zk.PermAll))
	}, z.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "create persistent node")
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "create persistent node path [%s] fail", path)
		return
	}

	return nil
}

// CreateTempNode create temp node
func (z *ZkClient) CreateTempNode(path string, data string) (err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	b, e := z.IsExists(path)
	if e != nil {
		err = errors.Wrap(e, "create temp node")
		return
	}
	if b {
		return nil
	}

	if err = z.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return
	}

	e = WarpperTimeout(func() {
		_, err = z.conn.Create(path, []byte(data), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	}, z.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "create temp node")
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "create tempnode path [%s] fail", path)
		return
	}

	return nil
}

// CreatePersistentSeqNode create persistent seq node
func (z *ZkClient) CreatePersistentSeqNode(path string, data string) (rpath string, err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	if err = z.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return
	}

	e := WarpperTimeout(func() {
		rpath, err = z.conn.Create(path, []byte(data), zk.FlagSequence, zk.WorldACL(zk.PermAll))
	}, z.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "create persistent seq node")
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "create persistent seq node path [%s] fail", path)
		return
	}

	return rpath, nil
}

// CreateTempSeqNode create temp seq node
func (z *ZkClient) CreateTempSeqNode(path string, data string) (rpath string, err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	if err = z.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return
	}

	e := WarpperTimeout(func() {
		rpath, err = z.conn.Create(path, []byte(data), zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	}, z.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "create temp seq node")
		return
	}
	if err != nil {
		err = errors.Wrapf(err, "create temp seq node path [%s] fail", path)
		return
	}

	return rpath, nil
}
