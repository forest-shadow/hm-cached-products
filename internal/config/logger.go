package config

import (
	"fmt"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return true
	default:
		return false
	}
}

func (l *LogLevel) UnmarshalText(text []byte) error {
	value := LogLevel(text)
	if !value.IsValid() {
		return fmt.Errorf("unsupported logger level %q", text)
	}

	*l = value

	return nil
}

type LogFormat string

const (
	LogFormatJSON    LogFormat = "json"
	LogFormatConsole LogFormat = "console"
)

func (f LogFormat) IsValid() bool {
	switch f {
	case LogFormatJSON, LogFormatConsole:
		return true
	default:
		return false
	}
}

func (f *LogFormat) UnmarshalText(text []byte) error {
	value := LogFormat(text)
	if !value.IsValid() {
		return fmt.Errorf("unsupported logger format %q", text)
	}

	*f = value

	return nil
}

func DefaultLogFormat(appEnv AppEnv) LogFormat {
	switch appEnv {
	case AppEnvLocal, AppEnvDevelopment:
		return LogFormatConsole
	default:
		return LogFormatJSON
	}
}

type LoggerConfig struct {
	Level  LogLevel  `mapstructure:"level"`
	Format LogFormat `mapstructure:"format"`
}

func (c LoggerConfig) WithDefaults(appEnv AppEnv) LoggerConfig {
	if c.Level == "" {
		c.Level = LogLevelDebug
	}

	if c.Format == "" {
		c.Format = DefaultLogFormat(appEnv)
	}

	return c
}

func (c LoggerConfig) Validate() error {
	if !c.Level.IsValid() {
		return fmt.Errorf("unsupported logger level %q", c.Level)
	}

	if !c.Format.IsValid() {
		return fmt.Errorf("unsupported logger format %q", c.Format)
	}

	return nil
}
