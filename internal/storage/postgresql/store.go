package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

const (
	MaxConns        = 10
	MinConns        = 1
	MaxConnLifetime = 30 * time.Minute
	MaxConnIdleTime = 5 * time.Minute
)

func Init(ctx context.Context, dsn string, migrationsPath string) (*Store, error) {
	if migrationsPath != "" {
		if err := Migrate(dsn, migrationsPath); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = MaxConns
	cfg.MinConns = MinConns
	cfg.MaxConnLifetime = MaxConnLifetime
	cfg.MaxConnIdleTime = MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.pool == nil {
		return fmt.Errorf("postgres: store is nil")
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return s.pool.Ping(pingCtx)
}
