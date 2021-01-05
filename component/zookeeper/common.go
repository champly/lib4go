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

// CompareSlice compare string slice is equal
func CompareSlice(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	if len(s1) == 0 && len(s2) == 0 {
		return true
	}

	m := make(map[string]struct{}, len(s1))
	for _, v := range s1 {
		m[v] = struct{}{}
	}
	for _, v := range s2 {
		if _, ok := m[v]; !ok {
			return false
		}
	}
	return true
}
