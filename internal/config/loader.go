// Package config provides shared settings and loading helpers for application configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Load reads and validates API settings from the working directory.
func Load() (Config, error) {
	return loadFromDir(".")
}

// loadFromDir keeps the source directory explicit so the real loading
// path can be exercised with isolated config files in tests.
func loadFromDir(dir string) (Config, error) {
	const op = "config.loader"

	appEnv, err := EnvironmentName()
	if err != nil {
		return Config{}, fmt.Errorf("%s: get env name: %w", op, err)
	}
	v := NewViper(dir, appEnv)

	setDefaults(v, appEnv)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("read env.%s.yml: %w", appEnv, err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(BuildDecodeHook())); err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate configuration: %w", err)
	}
	if cfg.App.Env != appEnv {
		return Config{}, fmt.Errorf("app.env %q must match APP_ENV %q", cfg.App.Env, appEnv)
	}
	return cfg, nil
}

// EnvironmentName returns the required APP_ENV value after validating it.
func EnvironmentName() (AppEnv, error) {
	env := AppEnv(os.Getenv("APP_ENV"))
	if env == "" {
		return "", errors.New("APP_ENV is required")
	}
	if !env.IsValid() {
		return "", fmt.Errorf("unsupported APP_ENV %q", env)
	}
	return env, nil
}

// NewViper configures YAML discovery and APP-prefixed environment overrides.
func NewViper(dir string, env AppEnv) *viper.Viper {
	v := viper.New()
	v.SetConfigName("env." + string(env))
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	return v
}
