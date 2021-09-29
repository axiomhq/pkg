package cmd

import "os"

// exitCode describes an application exit code.
type exitCode uint8

// exit the application with the code.
func (ec exitCode) exit() {
	os.Exit(int(ec))
}

// All available exit codes.
const (
	exitOK exitCode = iota
	exitInternal
	exitConfig
)
