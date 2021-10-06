package cmd

import (
	"os"

	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"
)

type config struct {
	axiomOptions             []axiom.Option
	loggerOptions            []zap.Option
	requiredEnvVars          []string
	exitSignals              []os.Signal
	validateAxiomCredentials bool
}
