package postgres

import (
	"context"
	"fmt"
	"reflect"
)

const platformOpsOwner = "platformops"

type PlatformOpsDatabase interface {
	Execute(context.Context, string, ...any) (CommandResult, error)
	Query(context.Context, string, ...any) (Rows, error)
	QueryRow(context.Context, string, ...any) Row
}

type PlatformOpsTransactions interface {
	Within(context.Context, TransactionMode, func(PlatformOpsDatabase) error) error
}

type platformOpsDatabase struct{ executor BoundExecutor }
type platformOpsTransactions struct{ pool *Pool }
type platformOpsBackupTransactions struct{ pool *Pool }

func (database platformOpsDatabase) Execute(
	ctx context.Context, statement string, arguments ...any,
) (CommandResult, error) {
	return database.executor.Execute(ctx, statement, arguments...)
}

func (database platformOpsDatabase) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func (database platformOpsDatabase) Query(
	ctx context.Context, statement string, arguments ...any,
) (Rows, error) {
	return database.executor.Query(ctx, statement, arguments...)
}

func (runner platformOpsTransactions) Within(
	ctx context.Context,
	mode TransactionMode,
	callback func(PlatformOpsDatabase) error,
) error {
	if callback == nil {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.BeginRuntimeTransaction(ctx, mode)
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
	if err := callback(database); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

func (runner platformOpsBackupTransactions) Within(
	ctx context.Context,
	mode TransactionMode,
	callback func(PlatformOpsDatabase) error,
) error {
	if callback == nil || runner.pool == nil ||
		runner.pool.profile != ProfileDatabaseBackupExecutor {
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
	if err := callback(database); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

func BindPlatformOpsStore[T any](
	pool *Pool,
	factory func(PlatformOpsTransactions) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() ||
		(pool.profile != ProfileAPI && pool.profile != ProfileWorker) ||
		!pool.runtimeCapable() || factory == nil {
		return zero, fmt.Errorf("platform operations store binding is unavailable")
	}
	runner := platformOpsTransactions{pool: pool}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}

func BindPlatformOpsBackupStore[T any](
	pool *Pool,
	factory func(PlatformOpsTransactions) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() ||
		pool.profile != ProfileDatabaseBackupExecutor ||
		!pool.profileTransactionCapable() || factory == nil {
		return zero, fmt.Errorf("platform operations backup binding is unavailable")
	}
	runner := platformOpsBackupTransactions{pool: pool}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
