package onboarding

import (
	"context"
	"fmt"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Pool interface {
	PostgresOwnerBindingPool() *platformpostgres.Pool
}

func NewUnitOfWork(handle Pool) (onboardingapp.ActivationUnitOfWork, error) {
	if handle == nil {
		return nil, errUnavailable
	}
	pool := handle.PostgresOwnerBindingPool()
	if pool == nil {
		return nil, errUnavailable
	}
	begin := func(ctx context.Context) (activationBoundary, error) {
		runtime, err := pool.BeginRuntimeTransaction(ctx, platformpostgres.TransactionCommand)
		if err != nil {
			return activationBoundary{}, fmt.Errorf("begin tenancy activation transaction: %w", err)
		}
		scope, err := platformtx.NewActivationScope(runtime)
		if err != nil {
			_ = runtime.Rollback()
			return activationBoundary{}, err
		}
		transaction, err := bindActivationTransaction(scope)
		if err != nil {
			_ = runtime.Rollback()
			return activationBoundary{}, err
		}
		return activationBoundary{transaction: transaction,
			commit: runtime.Commit, rollback: runtime.Rollback}, nil
	}
	return &activationUnitOfWork{begin: begin}, nil
}

func bindActivationTransaction(
	scope platformtx.ActivationScope,
) (*activationTransaction, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"tenancy",
		func(database platformpostgres.BoundExecutor) *postgresStore {
			return storeFromBoundExecutor(database)
		},
	)
	if err != nil {
		return nil, err
	}
	store, err := platformtx.BindActivation(scope, binder)
	if err != nil {
		return nil, err
	}
	result := &activationTransaction{store: store}
	result.bind = func(owner onboardingapp.ActivationOwner) (
		onboardingapp.ActivationOwnerExecutor, error,
	) {
		return bindActivationOwner(scope, owner)
	}
	result.bindSchema = func(ctx context.Context, schema string) error {
		return platformtx.BindActivationTenantSchema(scope, ctx, schema)
	}
	result.open.Store(true)
	return result, nil
}

func bindActivationOwner(
	scope platformtx.ActivationScope,
	owner onboardingapp.ActivationOwner,
) (onboardingapp.ActivationOwnerExecutor, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		string(owner),
		func(database platformpostgres.BoundExecutor) onboardingapp.ActivationOwnerExecutor {
			return ownerExecutorFromBound(database)
		},
	)
	if err != nil {
		return onboardingapp.ActivationOwnerExecutor{}, err
	}
	return platformtx.BindActivation(scope, binder)
}

func storeFromBoundExecutor(database platformpostgres.BoundExecutor) *postgresStore {
	return &postgresStore{queryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) Row {
		return database.QueryRow(ctx, statement, arguments...)
	},
		execute: func(ctx context.Context, statement string, arguments ...any) error {
			_, err := database.Execute(ctx, statement, arguments...)
			return err
		}}
}

func ownerExecutorFromBound(
	database platformpostgres.BoundExecutor,
) onboardingapp.ActivationOwnerExecutor {
	return onboardingapp.ActivationOwnerExecutor{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) onboardingapp.ActivationOwnerRow {
		return database.QueryRow(ctx, statement, arguments...)
	},
		Execute: func(ctx context.Context, statement string, arguments ...any) error {
			_, err := database.Execute(ctx, statement, arguments...)
			return err
		}}
}
