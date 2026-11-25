package logger_test

import (
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/forest-shadow/hm-cached-products/internal/config"
)

func TestViperUnmarshalConfigValidation(t *testing.T) {
	t.Parallel()

	type cfg struct {
		Logger config.LoggerConfig `mapstructure:"logger"`
	}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		v := viper.New()
		v.SetConfigType("yaml")

		if err := v.ReadConfig(strings.NewReader("logger:\n  level: info\n  format: json\n")); err != nil {
			t.Fatalf("read config: %v", err)
		}

		var got cfg
		if err := v.Unmarshal(&got, viper.DecodeHook(config.BuildDecodeHook())); err != nil {
			t.Fatalf("expected successful unmarshal, got error: %v", err)
		}
		if got.Logger.Level != config.LogLevelInfo || got.Logger.Format != config.LogFormatJSON {
			t.Fatalf("unexpected logger config: %+v", got.Logger)
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		t.Parallel()

		v := viper.New()
		v.SetConfigType("yaml")

		if err := v.ReadConfig(strings.NewReader("logger:\n  level: trace\n  format: json\n")); err != nil {
			t.Fatalf("read config: %v", err)
		}

		var got cfg
		if err := v.Unmarshal(&got, viper.DecodeHook(config.BuildDecodeHook())); err == nil {
			t.Fatal("expected unmarshal error for invalid logger level")
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()

		v := viper.New()
		v.SetConfigType("yaml")

		if err := v.ReadConfig(strings.NewReader("logger:\n  level: info\n  format: pretty\n")); err != nil {
			t.Fatalf("read config: %v", err)
		}

		var got cfg
		if err := v.Unmarshal(&got, viper.DecodeHook(config.BuildDecodeHook())); err == nil {
			t.Fatal("expected unmarshal error for invalid logger format")
		}
	})
}
