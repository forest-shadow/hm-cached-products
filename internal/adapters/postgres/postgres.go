package postgres

import (
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/forest-shadow/hm-cached-products/internal/config"
)

// Open creates and verifies a pool from settings validated during configuration loading.
// The context bounds startup verification; the DSN configures new connections.
func Open(ctx context.Context, cfg config.PostgresConfig) (*sqlx.DB, error) {
	const op = "postgres.Open"

	db, err := sqlx.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("%s: open database: %w", op, err)
	}

	db.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.Pool.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.Pool.ConnMaxLifetime)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%s: ping database: %w", op, err)
	}

	return db, nil
}
