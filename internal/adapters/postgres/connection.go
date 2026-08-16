package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection struct {
	pool *pgxpool.Pool
}

func NewConnection(ctx context.Context, databaseURL string) (*Connection, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse da URL do banco de dados: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("erro ao fazer ping no banco de dados: %w", err)
	}

	return &Connection{pool: pool}, nil
}

func (c *Connection) Pool() *pgxpool.Pool {
	return c.pool
}

func (c *Connection) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}
