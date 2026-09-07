package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDB inicializa el pool de conexiones a la base de datos PostgreSQL
func InitDB(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
