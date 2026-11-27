package config

import (
	"errors"
	"fmt"
	"time"
)

type PostgresConfig struct {
	DSN              string        `mapstructure:"dsn"`
	Pool           PoolConfig    `mapstructure:"db_pool"`
	StartupTimeout   time.Duration `mapstructure:"startup_timeout"`
	OperationTimeout time.Duration `mapstructure:"operation_timeout"`
}

func (c PostgresConfig) Validate() error {
	if c.DSN == "" {
		return errors.New("DSN is required")
	}
	if err := c.Pool.Validate(); err != nil {
		return fmt.Errorf("db_pool: %w", err)
	}
	// Zero disables idle connection retention or expiration limits, which the API requires.
	switch {
	case c.Pool.MaxIdleConns == 0:
		return errors.New("postgres.db_pool.max_idle_conns must be a positive integer")
	case c.Pool.ConnMaxLifetime == 0:
		return errors.New("postgres.db_pool.conn_max_lifetime must be a positive Go duration")
	case c.Pool.ConnMaxIdleTime == 0:
		return errors.New("postgres.db_pool.conn_max_idle_time must be a positive Go duration")
	case c.StartupTimeout <= 0:
		return errors.New("postgres.startup_timeout must be a positive Go duration")
	case c.OperationTimeout <= 0:
		return errors.New("postgres.operation_timeout must be a positive Go duration")
	default:
		return nil
	}
}

// PoolConfig describes one PostgreSQL connection pool owned by the application runtime.
type PoolConfig struct {
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

func (c PoolConfig) Validate() error {
	switch {
	case c.MaxOpenConns <= 0:
		return errors.New("MaxOpenConns must be positive")
	case c.MaxIdleConns < 0:
		return errors.New("MaxIdleConns must not be negative")
	case c.MaxIdleConns > c.MaxOpenConns:
		return errors.New("MaxIdleConns exceeds MaxOpenConns")
	case c.ConnMaxLifetime < 0:
		return errors.New("ConnMaxLifetime must not be negative")
	case c.ConnMaxIdleTime < 0:
		return errors.New("ConnMaxIdleTime must not be negative")
	default:
		return nil
	}
}
