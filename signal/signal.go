package signal

import (
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler signal
func SetupSignalHandler() chan struct{} {
	stop := make(chan struct{})

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(-1)
	}()

	return stop
}
