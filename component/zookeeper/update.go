package zookeeper

import "github.com/pkg/errors"

// Update update path value
func (zc *ZkClient) Update(path string, data string) (err error) {
	if b, e := zc.IsExists(path); !b || e != nil {
		return errors.Wrapf(e, "update path [%s] value fail, node is not exists", path)
	}

	conn, err := zc.getComplexConn()
	if err != nil {
		return err
	}

	e := WarpperTimeout(func() {
		_, err = conn.Set(path, []byte(data), -1)
	}, zc.execTimeout)
	if e != nil {
		err = errors.Wrap(e, "update path")
		return
	}

	if err != nil {
		err = errors.Wrapf(err, "update path [%s] fail", path)
		return
	}

	return nil
}
