package config

import (
	"errors"
	"time"
)

type HttpServerConfig struct {
	Addr              string        `mapstructure:"addr"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
}

func (c HttpServerConfig) Validate() error {
	if c.Addr == "" {
		return errors.New("http_server.addr is required")
	}
	switch {
	case c.ReadTimeout < 0:
		return errors.New("http_server.read_timeout must not be negative")
	case c.ReadHeaderTimeout < 0:
		return errors.New("http_server.read_header_timeout must not be negative")
	case c.WriteTimeout < 0:
		return errors.New("http_server.write_timeout must not be negative")
	case c.IdleTimeout < 0:
		return errors.New("http_server.idle_timeout must not be negative")
	case c.ShutdownTimeout <= 0:
		return errors.New("http_server.shutdown_timeout must be a positive Go duration")
	}
	return nil
}
