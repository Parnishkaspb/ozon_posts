package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	return &DB{Pool: pool}, nil
}

func (d *DB) Close() { d.Pool.Close() }

func (d *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := d.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("sql=%s args=%v: %w", sql, args, err)
	}
	return rows, nil
}
