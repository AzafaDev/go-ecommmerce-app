package repository

// Hand-written; sqlc generate doesn't touch this file (only db.go, models.go, querier.go, *.sql.go).

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx runs fn inside a transaction, committing on success or rolling back on any error.
func WithTx(ctx context.Context, dbPool *pgxpool.Pool, fn func(*Queries) error) error {
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(New(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
