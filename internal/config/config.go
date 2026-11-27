package config

import (
	"fmt"
)

// Config groups settings for the HTTP API by component.
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	HttpServer HttpServerConfig `mapstructure:"http_server"`
	Postgres   PostgresConfig   `mapstructure:"postgres"`
}

func (c Config) Validate() error {
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("app: %w", err)
	}
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	if err := c.HttpServer.Validate(); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	if err := c.Postgres.Validate(); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}

	return nil
}
