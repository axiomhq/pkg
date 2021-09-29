package http

import (
	"context"
	"net"
	"time"

	"go.uber.org/zap"
)

// An Option modifies the behaviour of a `Server`.
type Option func(s *Server) error

// WithBaseContext sets the `context.Context` that will server as base context
// for incoming requests.
func WithBaseContext(ctx context.Context) Option {
	return func(s *Server) error {
		s.srv.BaseContext = func(net.Listener) context.Context { return ctx }
		return nil
	}
}

// WithLogger sets the logger used for the underlying `http.Server`. The
// supplied logger will be named "http" and log at error level.
func WithLogger(logger *zap.Logger) Option {
	return func(s *Server) (err error) {
		s.srv.ErrorLog, err = zap.NewStdLogAt(logger.Named("http"), zap.ErrorLevel)
		return err
	}
}

// WithShutdownTimeout sets the timeout of the graceful shutdown.
func WithShutdownTimeout(dur time.Duration) Option {
	return func(s *Server) error {
		s.shutdownTimeout = dur
		return nil
	}
}
