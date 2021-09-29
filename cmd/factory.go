package cmd

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/axiomhq/axiom-go/axiom"
)

// Factory provides access to the logger, Axiom client, etc.
type Factory struct {
	// Cancel is the `context.CancelFunc` that cancels the application lifecycle
	// context.
	Cancel context.CancelFunc
	// Logger provides the application logger. The logger can be enhanced with
	// fields. By default it uses a `zap.NewProduction()` logger but this can
	// be changed to `zap.NewDevelopment()` by setting the "DEBUG" environment
	// variable to something else than "0".
	Logger func(...zapcore.Field) *zap.Logger
	// Axiom returns the Axiom client.
	Axiom func() *axiom.Client
}
