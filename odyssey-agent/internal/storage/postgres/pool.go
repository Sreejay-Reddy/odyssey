package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/config"
)

func NewPool(ctx context.Context, cfg config.PostgresConfig, dsn string,) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = cfg.PoolSize
	poolConfig.MinConns = cfg.PoolSize

	return pgxpool.NewWithConfig(ctx, poolConfig)
}