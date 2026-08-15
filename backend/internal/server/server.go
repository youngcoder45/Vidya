// Package server owns the HTTP server lifecycle and router construction.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/schoolos/backend/internal/deps"
)

// Server wraps the HTTP server.
type Server struct {
	http *http.Server
	log  *slog.Logger
}

// New builds the server with the given dependencies.
func New(d *deps.Deps) *Server {
	engine := newRouter(d)
	return &Server{
		http: &http.Server{
			Addr:              ":" + d.Cfg.Port,
			Handler:           engine,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		log: d.Log,
	}
}

// Run blocks until a shutdown signal is received, then drains gracefully.
func (s *Server) Run() error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("api listening", "addr", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	case sig := <-stop:
		s.log.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("server: graceful shutdown: %w", err)
	}
	s.log.Info("server stopped cleanly")
	return nil
}
