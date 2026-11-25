package config

import (
	"fmt"
)

// Config groups settings for the HTTP API by component.
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	HttpServer HttpServerConfig `mapstructure:"http_server"`
	Logger     LoggerConfig     `mapstructure:"logger"`
}

func (c Config) Validate() error {
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("app: %w", err)
	}
	if err := c.HttpServer.Validate(); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger: %w", err)
	}
	return nil
}
