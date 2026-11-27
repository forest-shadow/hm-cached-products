package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadHTTPTimeouts(t *testing.T) {
	const yamlTimeouts = "  read_timeout: 2s\n  read_header_timeout: 500ms\n  write_timeout: 3s\n  idle_timeout: 1m\n  shutdown_timeout: 4s\n"
	type testCase struct {
		name    string
		yaml    string
		vars    map[string]string
		want    HttpServerConfig
		wantErr string
	}
	tests := []testCase{
		{
			name: "YAML durations",
			yaml: yamlTimeouts,
			want: HttpServerConfig{
				Addr: ":8080", ReadTimeout: 2 * time.Second,
				ReadHeaderTimeout: 500 * time.Millisecond, WriteTimeout: 3 * time.Second,
				IdleTimeout: time.Minute, ShutdownTimeout: 4 * time.Second,
			},
		},
		{
			name: "environment overrides YAML",
			yaml: yamlTimeouts,
			vars: map[string]string{
				"APP_HTTP_SERVER_READ_TIMEOUT": "7s", "APP_HTTP_SERVER_READ_HEADER_TIMEOUT": "1s",
				"APP_HTTP_SERVER_WRITE_TIMEOUT": "8s", "APP_HTTP_SERVER_IDLE_TIMEOUT": "2m",
				"APP_HTTP_SERVER_SHUTDOWN_TIMEOUT": "9s",
			},
			want: HttpServerConfig{
				Addr: ":8080", ReadTimeout: 7 * time.Second, ReadHeaderTimeout: time.Second,
				WriteTimeout: 8 * time.Second, IdleTimeout: 2 * time.Minute, ShutdownTimeout: 9 * time.Second,
			},
		},
		{
			name: "environment sets durations absent from YAML",
			vars: map[string]string{
				"APP_HTTP_SERVER_READ_TIMEOUT": "7s", "APP_HTTP_SERVER_READ_HEADER_TIMEOUT": "1s",
				"APP_HTTP_SERVER_WRITE_TIMEOUT": "8s", "APP_HTTP_SERVER_IDLE_TIMEOUT": "2m",
				"APP_HTTP_SERVER_SHUTDOWN_TIMEOUT": "9s",
			},
			want: HttpServerConfig{
				Addr: ":8080", ReadTimeout: 7 * time.Second, ReadHeaderTimeout: time.Second,
				WriteTimeout: 8 * time.Second, IdleTimeout: 2 * time.Minute, ShutdownTimeout: 9 * time.Second,
			},
		},
		{
			name: "explicit zero overrides defaults",
			yaml: "  read_timeout: 0s\n  read_header_timeout: 0s\n  write_timeout: 0s\n  idle_timeout: 0s\n  shutdown_timeout: 1s\n",
			want: HttpServerConfig{Addr: ":8080", ShutdownTimeout: time.Second},
		},
		{name: "invalid duration", yaml: "  read_timeout: tomorrow\n", wantErr: "read_timeout"},
		{name: "zero shutdown timeout", yaml: "  shutdown_timeout: 0s\n", wantErr: "http_server.shutdown_timeout must be a positive Go duration"},
		{name: "invalid environment duration", vars: map[string]string{"APP_HTTP_SERVER_IDLE_TIMEOUT": "tomorrow"}, wantErr: "idle_timeout"},
	}
	for _, key := range []string{"read_timeout", "read_header_timeout", "write_timeout", "idle_timeout", "shutdown_timeout"} {
		tests = append(tests, testCase{
			name: "negative " + key, yaml: "  " + key + ": -1s\n", wantErr: "http_server." + key,
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv("APP_ENV", "local")
			for key, value := range tt.vars {
				t.Setenv(key, value)
			}
			dir := t.TempDir()
			data := "app:\n  name: test\n" + testPostgresYAML + "http_server:\n  addr: ':8080'\n" + tt.yaml
			if err := os.WriteFile(filepath.Join(dir, "env.local.yml"), []byte(data), 0o600); err != nil {
				t.Fatal(err)
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
			if got.HttpServer != tt.want {
				t.Fatalf("HTTP config = %+v, want %+v", got.HttpServer, tt.want)
			}
		})
	}
}
