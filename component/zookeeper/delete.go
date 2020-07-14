package zookeeper

import (
	"fmt"

	"github.com/pkg/errors"
)

// Delete delete node
func (z *ZkClient) Delete(path string) (err error) {
	if !z.isConnect {
		err = ErrClientDisConnect
		return
	}

	children, _, err := z.GetChildren(path)
	if err != nil {
		err = errors.Wrap(err, "delete node")
		return
	}

	for _, ch := range children {
		err = z.Delete(fmt.Sprintf("%s/%s", path, ch))
		if err != nil {
			return err
		}
	}

	e := WarpperTimeout(func() {
		err = z.conn.Delete(path, -1)
	}, z.execTimeout)
	if e != nil {
		err = errors.Wrap(err, "delete path")
		return
	}

	if err != nil {
		err = errors.Wrapf(err, "delete path [%s] fail", path)
		return
	}

	return nil
}
