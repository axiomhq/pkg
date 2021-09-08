package http

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"
)

// Server implements a http server which handles request using the configured
// `http.Handler`. It provides graceful shutdown capabilities using
// `context.Context` propagation.
type Server struct {
	lis net.Listener
	srv *http.Server

	doneCh chan struct{}
	errCh  chan error
}

// NewServer creates a new http server listening on the configured address and
// handling requests using the configured `http.Handler`. An error is returned
// if the address the server should listen on is already in use.
func NewServer(addr string, handler http.Handler) (*Server, error) {
	// Create tcp listener upfront to make sure the address to listen on is not
	// blocked.
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Create new http server.
	srv := &http.Server{
		Addr:     addr,
		Handler:  handler,
		ErrorLog: log.Default(),

		// Timeouts.
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 60,
	}

	return &Server{
		lis: lis,
		srv: srv,

		doneCh: make(chan struct{}),
		errCh:  make(chan error),
	}, nil
}

// ListenAddr returns the address the http server is listening on.
func (s *Server) ListenAddr() net.Addr {
	return s.lis.Addr()
}

// ListenError returns the receive-only channel which signals errors during http
// server startup.
func (s *Server) ListenError() <-chan error {
	return s.errCh
}

// Run starts the http server in a separate goroutine and pushes errors into the
// errors channel. This method is non-blocking. Use the `ListenError()` method
// to listen for errors which occure during startup and handle them accordingly.
func (s *Server) Run(ctx context.Context) {
	go func() {
		// Create the context to be used for the http servers base context for
		// incoming requests.
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()

		s.srv.BaseContext = func(net.Listener) context.Context { return ctx }
		s.srv.RegisterOnShutdown(cancel)

		// Serve on the configured listener.
		if err := s.srv.Serve(s.lis); err != http.ErrServerClosed {
			s.errCh <- err
		}
		close(s.errCh)
		close(s.doneCh)
	}()
}

// Shutdown stops the http server gracefully.
func (s *Server) Shutdown(ctx context.Context) error {
	// Shutdown the http server.
	s.srv.SetKeepAlivesEnabled(false)
	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}

	// Wait for the context to expire or a graceful shutdown signaled by the
	// done channel.
	select {
	case <-s.doneCh:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
