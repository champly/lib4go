package zookeeper

import (
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
)

// GetValue get path data
func (zc *ZkClient) GetValue(path string) (data string, version int32, err error) {
	var conn *complexConn
	conn, err = zc.getComplexConn()
	if err != nil {
		return
	}

	var stat *zk.Stat
	var value []byte
	e := WarpperTimeout(func() {
		value, stat, err = conn.Get(path)
	}, zc.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "get value function")
		return
	}

	if err != nil {
		err = errors.Wrapf(err, "get path [%s] value fail", path)
		return
	}

	version = stat.Version
	data = string(value)
	return
}

// GetChildren get children path
func (zc *ZkClient) GetChildren(path string) (children []string, version int32, err error) {
	var conn *complexConn
	conn, err = zc.getComplexConn()
	if err != nil {
		return
	}

	var stat *zk.Stat
	e := WarpperTimeout(func() {
		children, stat, err = conn.Children(path)
	}, zc.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "get children function")
		return
	}

	if err != nil {
		err = errors.Wrapf(err, "get path [%s] children fail", path)
		return
	}

	version = stat.Version
	return
}

// IsExists judge path is exists
func (zc *ZkClient) IsExists(path string) (b bool, err error) {
	var conn *complexConn
	conn, err = zc.getComplexConn()
	if err != nil {
		return
	}

	e := WarpperTimeout(func() {
		b, _, err = conn.Exists(path)
	}, zc.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "is exists")
		return
	}

	if err != nil {
		err = errors.Wrapf(err, "is exsts path [%s] fail", path)
		return
	}

	return
}
