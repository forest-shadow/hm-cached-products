package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testPostgresYAML = "postgres:\n  dsn: postgres://localhost/test\n  startup_timeout: 5s\n  operation_timeout: 5s\n  db_pool:\n    max_open_conns: 20\n    max_idle_conns: 5\n    conn_max_lifetime: 30m\n    conn_max_idle_time: 5m\n"

func TestLoadFromDir(t *testing.T) {
	defaultHTTP := HttpServerConfig{
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   10 * time.Second,
	}
	overriddenHTTP := defaultHTTP
	overriddenHTTP.Addr = ":9090"
	// PostgreSQL is now required by Config.Validate, even in HTTP config tests.
	postgresConfig := PostgresConfig{
		DSN: "postgres://localhost/test",
		Pool: PoolConfig{
			MaxOpenConns: 20, MaxIdleConns: 5,
			ConnMaxLifetime: 30 * time.Minute, ConnMaxIdleTime: 5 * time.Minute,
		},
		StartupTimeout: 5 * time.Second, OperationTimeout: 5 * time.Second,
	}
	tests := []struct {
		name    string
		env     string
		yaml    string
		vars    map[string]string
		want    Config
		wantErr string
	}{
		{
			name: "local defaults without app env in YAML",
			env:  "local",
			yaml: "app:\n  name: cached-products\n",
			want: Config{
				App:        AppConfig{Name: "cached-products", Env: AppEnvLocal},
				HttpServer: defaultHTTP,
				Logger:     LoggerConfig{Level: LogLevelDebug, Format: LogFormatConsole},
			},
		},
		{
			name: "production defaults",
			env:  "production",
			yaml: "app:\n  name: cached-products\n",
			vars: map[string]string{"APP_APP_NAME": "cached-products"},
			want: Config{
				App:        AppConfig{Name: "cached-products", Env: AppEnvProduction},
				HttpServer: defaultHTTP,
				Logger:     LoggerConfig{Level: LogLevelDebug, Format: LogFormatJSON},
			},
		},
		{
			name: "environment overrides YAML",
			env:  "local",
			yaml: "app:\n  name: yaml-name\n  env: local\nhttp_server:\n  addr: ':8081'\nlogger:\n  level: info\n  format: json\n",
			vars: map[string]string{
				"APP_APP_NAME":         "env-name",
				"APP_HTTP_SERVER_ADDR": ":9090",
				"APP_LOGGER_LEVEL":     "warn",
				"APP_LOGGER_FORMAT":    "console",
			},
			want: Config{
				App:        AppConfig{Name: "env-name", Env: AppEnvLocal},
				HttpServer: overriddenHTTP,
				Logger:     LoggerConfig{Level: LogLevelWarn, Format: LogFormatConsole},
			},
		},
		{name: "missing PostgreSQL without YAML", env: "production", vars: map[string]string{"APP_APP_NAME": "cached-products"}, wantErr: "DSN is required"},
		{name: "missing APP_ENV", wantErr: "APP_ENV is required"},
		{name: "invalid APP_ENV", env: "typo", wantErr: "unsupported APP_ENV"},
		{name: "missing app name", env: "local", wantErr: "app.name is required"},
		{name: "blank app name", env: "local", yaml: "app:\n  name: '   '\n", wantErr: "app.name is required"},
		{name: "invalid app env", env: "local", yaml: "app:\n  name: test\n  env: typo\n", wantErr: "unsupported app env"},
		{name: "mismatched app env", env: "local", yaml: "app:\n  name: test\n  env: production\n", wantErr: "must match APP_ENV"},
		{name: "invalid log level", env: "local", yaml: "app:\n  name: test\nlogger:\n  level: trace\n", wantErr: "unsupported logger level"},
		{name: "invalid log format", env: "local", yaml: "app:\n  name: test\nlogger:\n  format: xml\n", wantErr: "unsupported logger format"},
		{name: "invalid environment override", env: "local", vars: map[string]string{"APP_APP_NAME": "test", "APP_LOGGER_LEVEL": "trace"}, wantErr: "unsupported logger level"},
		{name: "empty HTTP address", env: "local", yaml: "app:\n  name: test\nhttp_server:\n  addr: ''\n", wantErr: "http_server.addr is required"},
		{name: "malformed YAML", env: "local", yaml: "app: [\n", wantErr: "read env.local.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Environment is process-wide, so these subtests must run sequentially.
			clearConfigEnv(t)
			t.Setenv("APP_ENV", tt.env)
			for key, value := range tt.vars {
				t.Setenv(key, value)
			}
			dir := t.TempDir()
			if tt.yaml != "" {
				if err := os.WriteFile(filepath.Join(dir, "env."+tt.env+".yml"), []byte(tt.yaml+"\n"+testPostgresYAML), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			got, err := loadFromDir(dir)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tt.want.Postgres = postgresConfig
			if got != tt.want {
				t.Fatalf("config = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestExampleConfig(t *testing.T) {
	clearConfigEnv(t)
	// Use the same example file that users copy into env.local.yml.
	data, err := os.ReadFile(filepath.Join("..", "..", "env.example.yml"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "env.local.yml"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV", "local")
	cfg, err := loadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.App.Name != "cached-products" || cfg.App.Env != AppEnvLocal {
		t.Fatalf("unexpected app config: %+v", cfg.App)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "APP_") {
			t.Setenv(key, "")
		}
	}
}
