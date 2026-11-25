package config

import (
	"github.com/spf13/viper"
)

func setDefaults(v *viper.Viper, appEnv AppEnv) {
	for key, value := range map[string]any{
		"app.name":         "",
		"app.env":          string(appEnv),
		"http_server.addr": ":8080",
		"logger.level":     string(LogLevelDebug),
		"logger.format":    string(DefaultLogFormat(appEnv)),
	} {
		v.SetDefault(key, value)
	}
}
