package postgres

import (
	"context"
	"fmt"
)

type IAMPostRestoreObserverDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

func (pool *Pool) WithinIAMPostRestoreObservation(
	ctx context.Context,
	callback func(IAMPostRestoreObserverDatabase) error,
) error {
	if pool == nil || pool.profile != ProfilePostRestoreObserver ||
		callback == nil {
		return fmt.Errorf("post-restore reactivation observation is unavailable")
	}
	transaction, err := pool.beginProfileTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	binder, err := NewOwnerBinder(
		"iam",
		func(executor BoundExecutor) IAMPostRestoreObserverDatabase {
			return executor
		},
	)
	if err != nil {
		return err
	}
	database, err := BindOwner(transaction, binder)
	if err != nil {
		return err
	}
	if err = callback(database); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}
