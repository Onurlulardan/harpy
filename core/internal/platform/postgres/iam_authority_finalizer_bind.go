package postgres

import (
	"context"
	"fmt"
	"reflect"
)

type IAMAuthorityFinalizerDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type IAMAuthorityFinalizerTransactions interface {
	Within(
		context.Context,
		TransactionMode,
		func(IAMAuthorityFinalizerDatabase) error,
	) error
}

type iamAuthorityFinalizerDatabase struct{ executor BoundExecutor }
type iamAuthorityFinalizerTransactions struct{ pool *Pool }

func (database iamAuthorityFinalizerDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func (runner iamAuthorityFinalizerTransactions) Within(
	ctx context.Context,
	mode TransactionMode,
	callback func(IAMAuthorityFinalizerDatabase) error,
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
		"iam",
		func(executor BoundExecutor) iamAuthorityFinalizerDatabase {
			return iamAuthorityFinalizerDatabase{executor: executor}
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

func BindIAMAuthorityFinalizer[T any](
	pool *Pool,
	factory func(IAMAuthorityFinalizerTransactions) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() || pool.profile != ProfileWorker ||
		!pool.runtimeCapable() || factory == nil {
		return zero, fmt.Errorf("IAM authority finalizer binding is unavailable")
	}
	runner := iamAuthorityFinalizerTransactions{pool: pool}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
