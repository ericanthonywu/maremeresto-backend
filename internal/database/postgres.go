package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ericanthonywu/maremereso-olga/backend/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const slowQueryThreshold = 500 * time.Millisecond

type queryStartKey struct{}

type queryStart struct {
	at  time.Time
	sql string
}

// queryTracer logs every query at debug level, failed queries at error level
// and queries slower than slowQueryThreshold at warn level.
type queryTracer struct{}

func (queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, queryStartKey{}, queryStart{at: time.Now(), sql: data.SQL})
}

func (queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value(queryStartKey{}).(queryStart)
	if !ok {
		return
	}
	elapsed := time.Since(start.at)
	attrs := []any{"sql", start.sql, "duration", elapsed, "command_tag", data.CommandTag.String()}

	switch {
	case data.Err != nil:
		slog.ErrorContext(ctx, "Database query failed", append(attrs, "err", data.Err)...)
	case elapsed >= slowQueryThreshold:
		slog.WarnContext(ctx, "Slow database query", attrs...)
	default:
		slog.DebugContext(ctx, "Database query executed", attrs...)
	}
}

func NewPool(cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		slog.Error("Failed to parse database config", "host", cfg.DBHost, "port", cfg.DBPort, "db", cfg.DBName, "err", err)
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Optimal pool settings for high concurrency
	poolConfig.MaxConns = 30
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	poolConfig.ConnConfig.Tracer = queryTracer{}

	slog.Debug("Connecting to PostgreSQL",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"db", cfg.DBName,
		"user", cfg.DBUser,
		"sslmode", cfg.DBSSLMode,
		"max_conns", poolConfig.MaxConns,
		"min_conns", poolConfig.MinConns,
	)
	startedAt := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		slog.Error("Failed to create connection pool", "host", cfg.DBHost, "port", cfg.DBPort, "db", cfg.DBName, "err", err)
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		slog.Error("Failed to ping database", "host", cfg.DBHost, "port", cfg.DBPort, "db", cfg.DBName, "elapsed", time.Since(startedAt), "err", err)
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL database", "db", cfg.DBName, "elapsed", time.Since(startedAt))
	return pool, nil
}
