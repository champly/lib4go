package zookeeper

import "github.com/pkg/errors"

// Update update path value
func (z *ZkClient) Update(path string, data string) (err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	if b, e := z.IsExists(path); !b || e != nil {
		err = errors.Wrapf(e, "update path [%s] value fail, node is not exists", path)
		return err
	}

	e := WarpperTimeout(func() {
		_, err = z.conn.Set(path, []byte(data), -1)
	}, z.execTimeout)
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
