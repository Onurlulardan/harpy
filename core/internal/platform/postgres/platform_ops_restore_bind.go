package postgres

import (
	"context"
	"fmt"
	"reflect"
)

type platformOpsRestoreTransactions struct{ pool *Pool }

func (runner platformOpsRestoreTransactions) Within(
	ctx context.Context,
	mode TransactionMode,
	callback func(PlatformOpsDatabase) error,
) error {
	if callback == nil || runner.pool == nil ||
		runner.pool.profile != ProfileDatabaseRestore {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.beginProfileTransaction(ctx, mode)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	binder, err := NewOwnerBinder(platformOpsOwner,
		func(executor BoundExecutor) platformOpsDatabase {
			return platformOpsDatabase{executor: executor}
		})
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

func BindPlatformOpsRestoreStore[T any](
	pool *Pool,
	factory func(PlatformOpsTransactions) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() ||
		pool.profile != ProfileDatabaseRestore ||
		!pool.profileTransactionCapable() || factory == nil {
		return zero, fmt.Errorf("platform operations restore binding is unavailable")
	}
	runner := platformOpsRestoreTransactions{pool: pool}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
