package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"

	"github.com/forest-shadow/hm-cached-products/internal/adapters/httpapi"
	"github.com/forest-shadow/hm-cached-products/internal/adapters/postgres"
	productsrepo "github.com/forest-shadow/hm-cached-products/internal/adapters/postgres/products"
	"github.com/forest-shadow/hm-cached-products/internal/application"
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

	db, productsRepo, err := newPostgresStorage(signalCtx, config.Postgres)
	if err != nil {
		return fmt.Errorf("new postgres storage: %w", err)
	}
	defer db.Close()

	server, err := newHttpServer(config.HttpServer, productsRepo, productsRepo, l)
	if err != nil {
		return fmt.Errorf("create HTTP handler: %w", err)
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ListenAndServe() }()
	l.Info("starting HTTP server", "addr", config.HttpServer.Addr)

	select {
	case <-signalCtx.Done():
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.HttpServer.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP Server: %w", err)
	}

	return nil
}

// newPostgresStorage transfers pool ownership to the caller only on success.
func newPostgresStorage(ctx context.Context, cfg config.PostgresConfig) (*sqlx.DB, *productsrepo.Repository, error) {
	startupCtx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	db, err := postgres.Open(startupCtx, config.PostgresConfig{
		DSN:  cfg.DSN,
		Pool: cfg.Pool,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open PostgreSQL: %w", err)
	}
	repository := productsrepo.New(db, cfg.OperationTimeout)
	return db, repository, nil
}

func newHttpServer(
	cfg config.HttpServerConfig,
	products application.ProductRepository,
	stats application.StatsReader,
	logger *logger.Logger,
) (*http.Server, error) {
	apiHandler, err := httpapi.New(logger, application.NewService(products, stats))
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           apiHandler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}, nil
}
