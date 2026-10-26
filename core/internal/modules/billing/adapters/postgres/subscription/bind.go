package subscription

import (
	"context"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func BindActor(
	scope platformtx.TenantScope,
	effects Effects,
) (subscriptionapp.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return boundStore(database, scope.TenantID(), scope.ActorID(), modeActor,
				effects, func(ctx context.Context) bool {
					return platformtx.ContextOwnsScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx) &&
						platformtx.RequireActive(scope, platformtx.Command) == nil
				})
		},
	)
	if err != nil {
		return nil, subscriptionapp.ErrConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, subscriptionapp.ErrConfiguration
	}
	return bound, nil
}

func BindQueries(scope platformtx.TenantScope) (subscriptionapp.CurrentStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *queryStore {
			return &queryStore{
				tenantID: scope.TenantID(),
				queryRow: func(ctx context.Context, statement string, arguments ...any) row {
					return database.QueryRow(ctx, statement, arguments...)
				},
				allowed: func(ctx context.Context) bool {
					return platformtx.ContextOwnsScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx) &&
						platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
				},
				notFound: platformpostgres.RowNotFound,
			}
		},
	)
	if err != nil {
		return nil, subscriptionapp.ErrQueryConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, subscriptionapp.ErrQueryConfiguration
	}
	return bound, nil
}

func BindProvider(
	scope platformtx.JobTenantScope,
	effects Effects,
) (subscriptionapp.Store, error) {
	return bindWorker(scope, effects, modeProvider)
}

func BindJob(
	scope platformtx.JobTenantScope,
	effects Effects,
) (subscriptionapp.Store, error) {
	return bindWorker(scope, effects, modeJob)
}

func NewBoundJobStore(
	tenantID uuid.UUID,
	database platformtx.JobBoundExecutor,
	effects Effects,
) (subscriptionapp.Store, error) {
	return NewJobStore(tenantID, DatabaseAccess{
		QueryRow: func(
			ctx context.Context, statement string, arguments ...any,
		) DatabaseRow {
			return database.QueryRow(ctx, statement, arguments...)
		},
		Execute: func(
			ctx context.Context, statement string, arguments ...any,
		) error {
			_, err := database.Execute(ctx, statement, arguments...)
			return err
		},
		Allowed: func(ctx context.Context) bool {
			return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx)
		},
		NotFound: platformpostgres.RowNotFound,
	}, effects)
}

func bindWorker(
	scope platformtx.JobTenantScope,
	effects Effects,
	mode bindingMode,
) (subscriptionapp.Store, error) {
	binder, err := platformtx.NewJobOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return boundStore(database, scope.TenantID(), uuid.Nil, mode, effects,
				func(ctx context.Context) bool {
					return platformtx.ContextOwnsJobScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx)
				})
		},
	)
	if err != nil {
		return nil, subscriptionapp.ErrConfiguration
	}
	bound, err := platformtx.BindJob(scope, binder)
	if err != nil || bound == nil {
		return nil, subscriptionapp.ErrConfiguration
	}
	return bound, nil
}

func boundStore(
	database platformpostgres.BoundExecutor,
	tenantID, actorID uuid.UUID,
	mode bindingMode,
	effects Effects,
	allowed scopeCheck,
) *store {
	return &store{
		tenantID: tenantID, actorID: actorID, mode: mode, effects: effects,
		queryRow: func(ctx context.Context, statement string, arguments ...any) row {
			return database.QueryRow(ctx, statement, arguments...)
		},
		execute: func(ctx context.Context, statement string, arguments ...any) error {
			_, err := database.Execute(ctx, statement, arguments...)
			return err
		},
		allowed: allowed, notFound: platformpostgres.RowNotFound,
	}
}
