package config

import "errors"

type HttpServerConfig struct {
	Addr string `mapstructure:"addr"`
}

func (c HttpServerConfig) Validate() error {
	if c.Addr == "" {
		return errors.New("http_server.addr is required")
	}
	return nil
}
