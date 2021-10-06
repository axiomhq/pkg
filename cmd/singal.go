package cmd

import (
	"os"
	"syscall"
)

// DefaultExitSignals are the default signals to catch and exit upon.
func DefaultExitSignals() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGHUP,
	}
}
