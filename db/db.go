package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalid = errors.New("db: invalid connection pool")

func New(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		wrapped := fmt.Errorf("%w: %v", ErrInvalid, err)
		return nil, wrapped
	}
	ping := pool.Ping(ctx)
	return pool, ping
}
