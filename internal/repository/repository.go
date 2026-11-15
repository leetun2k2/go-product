package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leetun2k2/go-bedrock/logx"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	pool   *pgxpool.Pool
	logger *logx.Logger
}

func New(pool *pgxpool.Pool, logger *logx.Logger) *Repository {
	return &Repository{
		pool:   pool,
		logger: logger,
	}
}

func (r *Repository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r *Repository) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *Repository) getExecutor(tx pgx.Tx) DBTX {
	if tx != nil {
		return tx
	}
	return r.pool
}
