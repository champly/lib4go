package zookeeper

import (
	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
)

// GetValue get path data
func (z *ZkClient) GetValue(path string) (data string, version int32, err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	var stat *zk.Stat
	var value []byte
	e := WarpperTimeout(func() {
		value, stat, err = z.conn.Get(path)
	}, z.execTimeout)
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
func (z *ZkClient) GetChildren(path string) (children []string, version int32, err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	var stat *zk.Stat
	e := WarpperTimeout(func() {
		children, stat, err = z.conn.Children(path)
	}, z.execTimeout)
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
func (z *ZkClient) IsExists(path string) (b bool, err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	e := WarpperTimeout(func() {
		b, _, err = z.conn.Exists(path)
	}, z.execTimeout)
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
