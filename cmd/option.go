package cmd

import (
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"
)

// An Option modifies the behaviour of the `Run()` function.
type Option func(c *config) error

// WithAxiomOptions sets the options used for creating the Axiom client.
func WithAxiomOptions(options ...axiom.Option) Option {
	return func(c *config) error {
		c.axiomOptions = options
		return nil
	}
}

// WithLoggerOptions sets the options used for creating the logger. If this
// option is not specified, the default logger options are specified by the
// `DefaultLoggerOptions()` function.
func WithLoggerOptions(options ...zap.Option) Option {
	return func(c *config) error {
		c.loggerOptions = options
		return nil
	}
}

// WithRequiredEnvVars sets the environment variables that are required to be
// set at application startup. Required environment variables must be set and
// not be empty. "AXIOM_URL", "AXIOM_TOKEN", "AXIOM_ORG_ID" and "DEBUG" are
// reserved.
func WithRequiredEnvVars(envVars ...string) Option {
	return func(c *config) error {
		c.requiredEnvVars = envVars
		return nil
	}
}

// WithSignals sets the signals that will cause the program to exit gracefully.
// If this option is not specified, the default signals are specified by the
// `DefaultSignals()` function.
func WithSignals(signals ...os.Signal) Option {
	return func(c *config) error {
		c.signals = signals
		return nil
	}
}

// WithValidateAxiomCredentials will validate the Axiom credentials at startup
// and fail the execution gracefully, if they are invalid.
func WithValidateAxiomCredentials() Option {
	return func(c *config) error {
		c.validateAxiomCredentials = true
		return nil
	}
}
