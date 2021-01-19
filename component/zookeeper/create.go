package zookeeper

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
)

// CreatePersistentNode create persistent node with data
func (zc *ZkClient) CreatePersistentNode(path string, data string) (err error) {
	conn, err := zc.getComplexConn()
	if err != nil {
		return err
	}

	if strings.EqualFold(path, "/") {
		return nil
	}

	if !strings.HasPrefix(path, "/") {
		err = fmt.Errorf("path should start with '/':%s", path)
		return
	}

	b, e := zc.IsExists(path)
	if e != nil {
		err = errors.Wrap(e, "create persistent node")
		return
	}
	if b {
		return nil
	}

	if err = zc.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return err
	}

	e = WarpperTimeout(func() {
		_, err = conn.Create(path, []byte(data), int32(0), zk.WorldACL(zk.PermAll))
	}, zc.execTimeout)
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
func (zc *ZkClient) CreateTempNode(path string, data string) (err error) {
	conn, err := zc.getComplexConn()
	if err != nil {
		return err
	}

	b, e := zc.IsExists(path)
	if e != nil {
		err = errors.Wrap(e, "create temp node")
		return
	}
	if b {
		return nil
	}

	if err = zc.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return
	}

	e = WarpperTimeout(func() {
		_, err = conn.Create(path, []byte(data), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	}, zc.execTimeout)
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
func (zc *ZkClient) CreatePersistentSeqNode(path string, data string) (rpath string, err error) {
	conn, err := zc.getComplexConn()
	if err != nil {
		return
	}

	if err = zc.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return
	}

	e := WarpperTimeout(func() {
		rpath, err = conn.Create(path, []byte(data), zk.FlagSequence, zk.WorldACL(zk.PermAll))
	}, zc.execTimeout)
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
func (zc *ZkClient) CreateTempSeqNode(path string, data string) (rpath string, err error) {
	conn, err := zc.getComplexConn()
	if err != nil {
		return
	}

	if err = zc.CreatePersistentNode(filepath.Dir(path), ""); err != nil {
		return
	}

	e := WarpperTimeout(func() {
		rpath, err = conn.Create(path, []byte(data), zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	}, zc.execTimeout)
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
