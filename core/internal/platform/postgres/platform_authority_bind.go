package postgres

import (
	"context"
	"fmt"
	"reflect"
)

const platformAuthorityOwner = "platformauthority"

type PlatformAuthorityDatabase interface {
	Execute(context.Context, string, ...any) (CommandResult, error)
	QueryRow(context.Context, string, ...any) Row
}

type PlatformAuthorityTransactions interface {
	Within(
		context.Context,
		TransactionMode,
		func(PlatformAuthorityDatabase) error,
	) error
}

type platformAuthorityDatabase struct{ executor BoundExecutor }
type platformAuthorityTransactions struct{ pool *Pool }

func (database platformAuthorityDatabase) Execute(
	ctx context.Context, statement string, arguments ...any,
) (CommandResult, error) {
	return database.executor.Execute(ctx, statement, arguments...)
}

func (database platformAuthorityDatabase) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func (runner platformAuthorityTransactions) Within(
	ctx context.Context,
	mode TransactionMode,
	callback func(PlatformAuthorityDatabase) error,
) error {
	if callback == nil {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.BeginRuntimeTransaction(ctx, mode)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	binder, err := NewOwnerBinder(
		platformAuthorityOwner,
		func(executor BoundExecutor) platformAuthorityDatabase {
			return platformAuthorityDatabase{executor: executor}
		},
	)
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

func BindPlatformAuthorityStore[T any](
	pool *Pool,
	factory func(PlatformAuthorityTransactions) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() ||
		(pool.profile != ProfileAPI && pool.profile != ProfileWorker &&
			pool.profile != ProfileOperatorInitialize) ||
		!pool.runtimeCapable() || factory == nil {
		return zero, fmt.Errorf("platform authority store binding is unavailable")
	}
	runner := platformAuthorityTransactions{pool: pool}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
