package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"sipasdos/config"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := (&url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			config.GetEnv("DB_USER", "postgres"),
			config.GetEnv("DB_PASSWORD", ""),
		),
		Host:     config.GetEnv("DB_HOST", "localhost") + ":" + config.GetEnv("DB_PORT", "5432"),
		Path:     "/" + config.GetEnv("DB_NAME", "sipasdos"),
		RawQuery: "sslmode=" + config.GetEnv("DB_SSLMODE", "disable"),
	}).String()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi database tidak valid: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database tidak dapat dihubungi: %w", err)
	}

	return pool, nil
}
