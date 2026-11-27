package postgres

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/forest-shadow/hm-cached-products/internal/config"
)

func TestOpenConnectionAndStartTimeouts(t *testing.T) {
	t.Setenv("PGCONNECT_TIMEOUT", "10")
	for _, tc := range []struct {
		name               string
		connectTimeout     string
		startupTimeout     time.Duration
		wantStartupTimeout bool
	}{
		{"driver timeout from dsn", "1", 3 * time.Second, false},
		{"startup timeout expires before connect timeoute", "10", 100 * time.Millisecond, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate a PostgreSQL server that accepts TCP connections but never
			// responds to the startup message, forcing the client to wait for a timeout.
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			go func() {
				defer close(done)
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
				_, _ = io.Copy(io.Discard, conn)
			}()
			t.Cleanup(func() {
				_ = listener.Close()
				<-done
			})

			ctx, cancel := context.WithTimeout(context.Background(), tc.startupTimeout)
			defer cancel()
			db, err := Open(ctx, config.PostgresConfig{
				DSN:  "postgres://test:test@" + listener.Addr().String() + "/test?sslmode=disable&connect_timeout=" + tc.connectTimeout,
				Pool: config.PoolConfig{MaxOpenConns: 1, MaxIdleConns: 1},
			})
			// A non-nil pool does not imply a working database connection.
			// Open must close the pool and return nil when startup verification fails.
			if db != nil {
				_ = db.Close()
				t.Fatal("Open returned a pool after an incomplete handshake")
			}
			if !pgconn.Timeout(err) {
				t.Fatalf("Open error = %v, want connection timeout", err)
			}
			if tc.wantStartupTimeout {
				if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
					t.Fatalf("startup context error = %v, want deadline exceeded", ctx.Err())
				}
			} else if ctx.Err() != nil {
				t.Fatalf("driver did not stop before startup deadline: %v", ctx.Err())
			}
		})
	}
}

func TestOpenRejectsInvalidDSNTimeout(t *testing.T) {
	for _, timeout := range []string{"tomorrow", "5s", "-1"} {
		t.Run(timeout, func(t *testing.T) {
			db, err := Open(context.Background(), config.PostgresConfig{
				DSN:  "postgres://localhost/test?connect_timeout=" + timeout,
				Pool: config.PoolConfig{MaxOpenConns: 1, MaxIdleConns: 1},
			})
			if db != nil {
				_ = db.Close()
				t.Fatal("Open returned a pool for an invalid connection timeout")
			}
			var parseErr *pgconn.ParseConfigError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Open error = %v, want DSN parsing error", err)
			}
		})
	}
}