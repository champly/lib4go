package cmd

import (
	"testing"
	"time"
)

func TestExecuteWithTimeout(t *testing.T) {
	outStr, err := ExecuteWithTimeout("cat /proc/cpuinfo | grep 'core id' | wc -l", time.Second*10)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(outStr)

	// os: process already finished
	outStr, err = ExecuteWithTimeout("sleep 3", time.Second*1)
	if err == nil {
		t.Error(outStr)
		return
	}
	t.Log(err)

	outStr, err = ExecuteWithTimeout("adfadf", time.Millisecond*200)
	if err == nil {
		t.Error(outStr)
		return
	}
	t.Log(err)
}
