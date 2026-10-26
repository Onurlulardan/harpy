package postgres

import (
	"context"
	"fmt"
)

type IAMInitialAdminSetupReissueDatabase interface {
	QueryRow(context.Context, string, ...any) Row
}

type iamInitialAdminSetupReissueDatabase struct{ executor BoundExecutor }

func (database iamInitialAdminSetupReissueDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return database.executor.QueryRow(ctx, statement, arguments...)
}

func RunIAMInitialAdminSetupReissue(
	ctx context.Context,
	pool *Pool,
	schemaName string,
	callback func(IAMInitialAdminSetupReissueDatabase) error,
) error {
	if ctx == nil || ctx.Err() != nil || pool == nil || pool.closed.Load() ||
		pool.profile != ProfileAPI || !pool.runtimeCapable() || callback == nil {
		return ErrAPIStoreProfile
	}
	transaction, err := pool.BeginRuntimeTransaction(ctx, TransactionCommand)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	if err = transaction.BindActivationTenantSearchPath(ctx, schemaName); err != nil {
		return err
	}
	binder, err := NewOwnerBinder("iam",
		func(executor BoundExecutor) IAMInitialAdminSetupReissueDatabase {
			return iamInitialAdminSetupReissueDatabase{executor: executor}
		})
	if err != nil {
		return fmt.Errorf("build initial admin setup reissue binder: %w", err)
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
