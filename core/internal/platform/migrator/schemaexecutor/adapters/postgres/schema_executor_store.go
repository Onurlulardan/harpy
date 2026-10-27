package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type schemaExecutorStore struct {
	pool           *pgxpool.Pool
	cleanupTimeout time.Duration
}

func schemaExecutorOperationContext(
	parent context.Context,
) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, schemaexecutor.DatabaseOperationTimeout)
}

func (store *schemaExecutorStore) withTransaction(
	ctx context.Context,
	apply func(pgx.Tx) error,
) error {
	transaction, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer store.rollbackTransaction(transaction)
	if err := apply(transaction); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

func (store *schemaExecutorStore) rollbackTransaction(transaction pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), store.cleanupTimeout)
	defer cancel()
	_ = transaction.Rollback(ctx)
}

var _ schemaexecutor.Persistence = (*schemaExecutorStore)(nil)
