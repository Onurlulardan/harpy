package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func withDeadLetterTransaction[T any](
	ctx context.Context,
	repository *deadLetterRepository,
	operation func(pgx.Tx) (T, error),
) (result T, resultErr error) {
	if ctx == nil || !repository.valid() || operation == nil {
		return result, ErrDeadLetterStoreUnavailable
	}
	transaction, err := repository.database.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return result, fmt.Errorf("begin dead-letter transaction: %w", err)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		cleanup, cancel := context.WithTimeout(context.Background(), repository.cleanup)
		defer cancel()
		resultErr = errors.Join(resultErr, transaction.Rollback(cleanup))
	}()
	result, resultErr = operation(transaction)
	if resultErr != nil {
		return result, resultErr
	}
	if err := transaction.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit dead-letter transaction: %w", err)
	}
	committed = true
	return result, nil
}
