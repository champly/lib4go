package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// ExecuteWithTimeout execute with timeout
func ExecuteWithTimeout(cmd string, t time.Duration) (result string, err error) {
	var out, e bytes.Buffer
	command := exec.Command("bash", "-c", cmd)
	command.Stdout = &out
	command.Stderr = &e

	if err = command.Start(); err != nil {
		return
	}

	done := make(chan error)
	go func() {
		done <- command.Wait()
	}()

	select {
	case <-time.After(t):
		// exec timeout
		command.Process.Signal(syscall.SIGINT)
		time.Sleep(time.Second)
		err = errors.New("execute command timeout")
		if ee := command.Process.Kill(); ee != nil && !strings.EqualFold("os: process already finished", ee.Error()) {
			err = fmt.Errorf("%s, kill command error:%s", err, ee.Error())
		}
	case err = <-done:
		result = trimOutput(out)
		if e.String() != "" {
			err = fmt.Errorf(e.String())
		}
	}
	return
}

func trimOutput(buf bytes.Buffer) string {
	return strings.TrimSpace(string(bytes.TrimRight(buf.Bytes(), "\x00")))
}
