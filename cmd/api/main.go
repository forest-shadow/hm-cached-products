package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/forest-shadow/hm-cached-products/internal/adapters/httpapi"
	"github.com/forest-shadow/hm-cached-products/internal/config"
	"github.com/forest-shadow/hm-cached-products/internal/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	config, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	l, err := logger.New(logger.Config{App: config.App, Logger: config.Logger})
	if err != nil {
		return fmt.Errorf("new logger: %w", err)
	}

	server := &http.Server{
		Addr:              config.HttpServer.Addr,
		Handler:           httpapi.New(l),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ListenAndServe() }()
	l.Info("starting HTTP server", logger.F("addr", config.HttpServer.Addr))

	select {
	case <-signalCtx.Done():
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP Server: %w", err)
	}

	return nil
}
