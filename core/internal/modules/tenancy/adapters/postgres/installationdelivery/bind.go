package installationdelivery

import (
	"context"
	"fmt"

	installationdelivery "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/installationdelivery"
	onboarding "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type commandStore struct{ pool *platformpostgres.Pool }

func NewCommandStore(pool *platformpostgres.Pool) (installationdelivery.Store, error) {
	if pool == nil {
		return nil, installationdelivery.ErrUnavailable
	}
	return &commandStore{pool: pool}, nil
}

func (storage *commandStore) open(ctx context.Context) (
	*platformpostgres.RuntimeTransaction, *store, error,
) {
	if storage == nil || storage.pool == nil || ctx == nil || ctx.Err() != nil {
		return nil, nil, installationdelivery.ErrUnavailable
	}
	runtime, err := storage.pool.BeginRuntimeTransaction(ctx, platformpostgres.TransactionCommand)
	if err != nil {
		return nil, nil, err
	}
	scope, err := platformtx.NewCredentialDeliveryScope(runtime)
	if err != nil {
		_ = runtime.Rollback()
		return nil, nil, err
	}
	binder, err := platformpostgres.NewOwnerBinder("tenancy", newStore)
	if err != nil {
		_ = runtime.Rollback()
		return nil, nil, err
	}
	bound, err := platformtx.BindCredentialDelivery(scope, binder)
	if err != nil {
		_ = runtime.Rollback()
		return nil, nil, err
	}
	return runtime, bound, nil
}

func NewRecoveryUnitOfWork(
	pool *platformpostgres.Pool,
) (onboarding.RecoveryUnitOfWork, error) {
	if pool == nil {
		return nil, errRecoveryUnavailable
	}
	return &recoveryUnit{begin: func(ctx context.Context) (recoveryBoundary, error) {
		runtime, err := pool.BeginRuntimeTransaction(ctx, platformpostgres.TransactionCommand)
		if err != nil {
			return recoveryBoundary{}, fmt.Errorf("begin installation recovery: %w", err)
		}
		scope, err := platformtx.NewCredentialDeliveryScope(runtime)
		if err != nil {
			_ = runtime.Rollback()
			return recoveryBoundary{}, err
		}
		transaction, err := bindRecoveryTransaction(scope)
		if err != nil {
			_ = runtime.Rollback()
			return recoveryBoundary{}, err
		}
		return recoveryBoundary{transaction: transaction,
			commit: runtime.Commit, rollback: runtime.Rollback}, nil
	}}, nil
}

func bindRecoveryTransaction(
	scope platformtx.CredentialDeliveryScope,
) (*recoveryTransaction, error) {
	tenancyBinder, err := platformpostgres.NewOwnerBinder("tenancy", newStore)
	if err != nil {
		return nil, err
	}
	storage, err := platformtx.BindCredentialDelivery(scope, tenancyBinder)
	if err != nil {
		return nil, err
	}
	result := &recoveryTransaction{store: storage}
	result.bindSchema = func(ctx context.Context, schema string) error {
		return platformtx.BindCredentialDeliveryTenantSchema(scope, ctx, schema)
	}
	result.bindIAM = func() (onboarding.RecoveryOwnerExecutor, error) {
		binder, buildErr := platformpostgres.NewOwnerBinder("iam", func(
			database platformpostgres.BoundExecutor,
		) onboarding.RecoveryOwnerExecutor {
			return onboarding.RecoveryOwnerExecutor{QueryRow: func(
				ctx context.Context, statement string, arguments ...any,
			) onboarding.RecoveryRow {
				return database.QueryRow(ctx, statement, arguments...)
			}}
		})
		if buildErr != nil {
			return onboarding.RecoveryOwnerExecutor{}, buildErr
		}
		return platformtx.BindCredentialDelivery(scope, binder)
	}
	result.open.Store(true)
	return result, nil
}

func newStore(database platformpostgres.BoundExecutor) *store {
	return &store{queryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) Row {
		return database.QueryRow(ctx, statement, arguments...)
	}}
}

func commitCommand(
	ctx context.Context, runtime *platformpostgres.RuntimeTransaction,
) error {
	if err := runtime.Commit(ctx); err != nil {
		return fmt.Errorf("commit installation delivery command: %w", err)
	}
	return nil
}

var _ installationdelivery.Store = (*commandStore)(nil)
