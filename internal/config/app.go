package config

import (
	"errors"
	"fmt"
	"strings"
)

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  AppEnv `mapstructure:"env"`
}

func (a AppConfig) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return errors.New("app.name is required")
	}

	if !a.Env.IsValid() {
		return fmt.Errorf("unsupported app env %q", a.Env)
	}

	return nil
}
