package postgres

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

const iamPostRestoreOwner = "iam"

type IAMPostRestoreDatabase interface {
	Execute(context.Context, string, ...any) (CommandResult, error)
	Query(context.Context, string, ...any) (Rows, error)
	QueryRow(context.Context, string, ...any) Row
}

type IAMPostRestoreTransactions interface {
	WithinGlobal(context.Context, TransactionMode, func(IAMPostRestoreDatabase) error) error
	WithinTenant(context.Context, TransactionMode, uuid.UUID,
		func(IAMPostRestoreDatabase) error) error
}

type iamPostRestoreTransactions struct{ pool *Pool }
type iamPostRestoreDatabase struct{ executor BoundExecutor }

func (database iamPostRestoreDatabase) Execute(
	ctx context.Context, statement string, arguments ...any,
) (CommandResult, error) {
	return database.executor.Execute(ctx, statement, arguments...)
}

func (database iamPostRestoreDatabase) Query(
	ctx context.Context, statement string, arguments ...any,
) (Rows, error) {
	return database.executor.Query(ctx, statement, arguments...)
}

func (database iamPostRestoreDatabase) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func (runner iamPostRestoreTransactions) WithinGlobal(
	ctx context.Context, mode TransactionMode,
	callback func(IAMPostRestoreDatabase) error,
) error {
	return runner.within(ctx, mode, uuid.Nil, false, callback)
}

func (runner iamPostRestoreTransactions) WithinTenant(
	ctx context.Context, mode TransactionMode, tenantID uuid.UUID,
	callback func(IAMPostRestoreDatabase) error,
) error {
	return runner.within(ctx, mode, tenantID, true, callback)
}

func (runner iamPostRestoreTransactions) within(
	ctx context.Context, mode TransactionMode, tenantID uuid.UUID, tenant bool,
	callback func(IAMPostRestoreDatabase) error,
) error {
	if callback == nil || (tenant && tenantID == uuid.Nil) {
		return ErrOwnerBindingInvalid
	}
	transaction, err := runner.pool.BeginRuntimeTransaction(ctx, mode)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	if tenant {
		record, resolveErr := transaction.ResolveTenant(ctx, tenantID)
		if resolveErr != nil {
			return resolveErr
		}
		if err = transaction.BindTenantSearchPath(ctx, record); err != nil {
			return err
		}
		if err = transaction.VerifyTenantIdentity(ctx, tenantID, record); err != nil {
			return err
		}
	}
	binder, err := NewOwnerBinder(iamPostRestoreOwner,
		func(executor BoundExecutor) iamPostRestoreDatabase {
			return iamPostRestoreDatabase{executor: executor}
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

func BindIAMPostRestoreStore[T any](
	pool *Pool, factory func(IAMPostRestoreTransactions) (T, error),
) (T, error) {
	var zero T
	if pool == nil || pool.closed.Load() || !pool.runtimeCapable() || factory == nil ||
		(pool.profile != ProfileAPI && pool.profile != ProfileOperatorIdentityReconcile) {
		return zero, fmt.Errorf("IAM post-restore store binding is unavailable")
	}
	runner := iamPostRestoreTransactions{pool: pool}
	result, err := factory(runner)
	if err != nil {
		return zero, err
	}
	if reflect.TypeOf(any(result)) == reflect.TypeOf(runner) {
		return zero, ErrOwnerBindingInvalid
	}
	return result, nil
}
