package postgres

import (
	"context"
	"fmt"
	"reflect"
)

const tenancyRecoveryOwner = "tenancy"

type TenancyRecoveryDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type TenancyRecoveryTransactions interface {
	Within(context.Context, TransactionMode, func(TenancyRecoveryDatabase) error) error
}

type tenancyRecoveryTransactions struct{ pool *Pool }
type tenancyRecoveryDatabase struct{ executor BoundExecutor }

func (database tenancyRecoveryDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func (runner tenancyRecoveryTransactions) Within(
	ctx context.Context,
	mode TransactionMode,
	callback func(TenancyRecoveryDatabase) error,
) error {
	if callback == nil {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.beginProfileTransaction(ctx, mode)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	binder, err := NewOwnerBinder(tenancyRecoveryOwner,
		func(executor BoundExecutor) tenancyRecoveryDatabase {
			return tenancyRecoveryDatabase{executor: executor}
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

func BindTenancyRecoveryStore[T any](
	pool *Pool,
	factory func(TenancyRecoveryTransactions) (T, error),
) (T, error) {
	var zero T
	if !recoveryProfile(pool) || factory == nil {
		return zero, fmt.Errorf("tenancy recovery store binding is unavailable")
	}
	runner := tenancyRecoveryTransactions{pool: pool}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}

func recoveryProfile(pool *Pool) bool {
	return pool != nil && !pool.closed.Load() && pool.profileTransactionCapable() &&
		(pool.profile == ProfileDatabaseBackupExecutor ||
			pool.profile == ProfileDatabaseRestore)
}
