package zookeeper

import (
	"time"
)

// WarpperTimeout exec function with timeout
func WarpperTimeout(f func(), timeout time.Duration) error {

	ch := make(chan struct{})

	go func() {
		defer close(ch)
		f()
	}()

	select {
	case <-time.After(timeout):
		return ErrExecTimeout
	case <-ch:
		return nil
	}
}
