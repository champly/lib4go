package zookeeper

import (
	"fmt"

	"github.com/pkg/errors"
)

// Delete delete node
func (zc *ZkClient) Delete(path string) (err error) {
	conn, err := zc.getComplexConn()
	if err != nil {
		return err
	}

	children, _, err := zc.GetChildren(path)
	if err != nil {
		err = errors.Wrap(err, "delete node")
		return
	}

	for _, ch := range children {
		err = zc.Delete(fmt.Sprintf("%s/%s", path, ch))
		if err != nil {
			return err
		}
	}

	e := WarpperTimeout(func() {
		err = conn.Delete(path, -1)
	}, zc.execTimeout)
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
