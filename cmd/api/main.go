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
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "load configuration:", err)
		os.Exit(1)
	}

	log, err := logger.New(logger.Config{App: cfg.App, Logger: cfg.Logger})
	if err != nil {
		fmt.Fprintln(os.Stderr, "create logger:", err)
		os.Exit(1)
	}

	err = run(cfg, log)
	if err != nil {
		log.Error("application failed", "error", err)
	}

	_ = log.Sync()

	if err != nil {
		os.Exit(1)
	}
}

func run(cfg config.Config, log *logger.Logger) error {
	const op = "api.run"

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, productsRepo, err := newPostgresStorage(signalCtx, cfg.Postgres)
	if err != nil {
		return fmt.Errorf("%s: new postgres storage: %w", op, err)
	}
	defer db.Close()

	server, err := newHttpServer(cfg.HttpServer, productsRepo, productsRepo, log)
	if err != nil {
		return fmt.Errorf("%s: create HTTP handler: %w", op, err)
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ListenAndServe() }()
	log.Info("starting HTTP server", "addr", cfg.HttpServer.Addr)

	select {
	case <-signalCtx.Done():
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("%s: serve HTTP: %w", op, err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HttpServer.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("%s: shutdown HTTP server: %w", op, err)
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
