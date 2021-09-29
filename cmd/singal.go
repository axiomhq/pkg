package cmd

import (
	"os"
	"syscall"
)

// DefaultSignals are the default signals to catch and act upon.
func DefaultSignals() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGHUP,
	}
}
