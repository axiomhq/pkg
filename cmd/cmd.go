package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/axiomhq/axiom-go/axiom"
	"go.uber.org/zap"

	"github.com/axiomhq/pkg/version"
)

// RunFunc is implemented by the main packages and passed to the `Run` function
// which takes care of signal handling, loading the runtime configuration and
// setting up logging, the Axiom client, etc. It must block until the context is
// marked done. Errors returned from the `RunFunc` should be created using the
// `Error()` function.
type RunFunc func(context.Context, *zap.Logger, *axiom.Client) error

// Run the named app with the given `RunFunc`. Additionally, options can be
// passed to configure the behaviour of the bootstrapping process.
func Run(appName string, fn RunFunc, options ...Option) {
	if code := run(appName, fn, options...); code != exitOK {
		code.exit()
	}
}

func run(appName string, fn RunFunc, options ...Option) exitCode {
	// Setup the default config and apply the supplied options.
	cfg := &config{
		loggerOptions: DefaultLoggerOptions(),
		exitSignals:   DefaultExitSignals(),
	}
	for _, option := range options {
		if err := option(cfg); err != nil {
			return exitConfig
		}
	}

	// Set up logger.
	var (
		logger *zap.Logger
		err    error
	)

	if v, _ := strconv.ParseBool(os.Getenv("DEBUG")); v {
		logger, err = zap.NewDevelopment(cfg.loggerOptions...)
	} else {
		logger, err = zap.NewProduction(cfg.loggerOptions...)
	}
	if err != nil {
		log.Printf("failed to create logger: %v", err)
		return exitConfig
	}
	defer func() {
		logger.Warn("stopped")

		// HINT(lukasmalkmus): Ignore error because of
		// https://github.com/uber-go/zap/issues/880.
		_ = logger.Sync()
	}()

	// Add application name to the logger
	logger = logger.Named(appName)

	// Log version information.
	logger.Info("starting",
		zap.String("release", version.Release()),
		zap.String("revision", version.Revision()),
		zap.String("build_date", version.BuildDateString()),
		zap.String("build_user", version.BuildUser()),
		zap.String("go_version", version.GoVersion()),
	)

	// Make sure the required environment variables are set.
	for _, env := range cfg.requiredEnvVars {
		if os.Getenv(env) == "" {
			logger.Error("missing environment variable", zap.String("name", env))
			return exitConfig
		}
	}

	// Listen for termination signals.
	ctx, cancel := signal.NotifyContext(context.Background(), cfg.exitSignals...)
	defer cancel()

	// Create the Axiom client.
	client, err := axiom.NewClient(cfg.axiomOptions...)
	if err != nil {
		logger.Error("create axiom client", zap.Error(err))
		return exitConfig
	}

	// If enabled, validate the credentials of the Axiom client.
	if cfg.validateAxiomCredentials {
		if err = client.ValidateCredentials(ctx); err != nil {
			logger.Error("validate axiom credentials", zap.Error(err))
			return exitConfig
		}
	}

	logger.Info("started")

	// Call the actual `RunFunc`. If the returned error was composed using
	// `cmd.Error()`, it can be logged properly. If not, logging the error is
	// done as well but with less context to it.
	if err = fn(ctx, logger, client); err != nil {
		if mainErr, ok := err.(*mainFuncError); ok {
			logger.Error(mainErr.msg, mainErr.Fields()...)
		} else {
			msg := fmt.Sprintf("%s.RunFunc", appName)
			logger.Error(msg, zap.Error(err))
		}
		return exitInternal
	}

	return exitOK
}
