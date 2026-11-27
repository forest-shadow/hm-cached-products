package config

import (
	"time"

	"github.com/spf13/viper"
)

func setDefaults(v *viper.Viper, appEnv AppEnv) {
	for key, value := range map[string]any{
		"app.name":                        "",
		"app.env":                         string(appEnv),
		"http_server.addr":                ":8080",
		"http_server.read_timeout":        time.Duration(0),
		"http_server.read_header_timeout": 5 * time.Second,
		"http_server.write_timeout":       time.Duration(0),
		"http_server.idle_timeout":        60 * time.Second,
		"http_server.shutdown_timeout":    10 * time.Second,
		"logger.level":                    string(LogLevelDebug),
		"logger.format":                   string(DefaultLogFormat(appEnv)),
	} {
		v.SetDefault(key, value)
	}
}
