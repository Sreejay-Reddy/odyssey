package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/config"
)

type Writer struct {
	pool *pgxpool.Pool
	cfg  config.Config
}

func New(pool *pgxpool.Pool, cfg config.Config) *Writer {
	return &Writer{
		pool: pool,
		cfg:  cfg,
	}
}