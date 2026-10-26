package usage

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func Bind(scope platformtx.TenantScope, effects Effects) (usageapp.Store, error) {
	if effects == nil {
		return nil, usageapp.ErrConfiguration
	}
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				tenantID: scope.TenantID(), actorID: scope.ActorID(),
				queryRow: func(ctx context.Context, sql string, arguments ...any) row {
					return database.QueryRow(ctx, sql, arguments...)
				},
				execute: func(ctx context.Context, sql string, arguments ...any) error {
					_, err := database.Execute(ctx, sql, arguments...)
					return err
				},
				notFound: platformpostgres.RowNotFound,
				effects:  effects,
				allowed: func(ctx context.Context) bool {
					return platformtx.ContextOwnsScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx) &&
						platformtx.RequireActive(scope, platformtx.Command) == nil
				},
			}
		},
	)
	if err != nil {
		return nil, usageapp.ErrConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, usageapp.ErrConfiguration
	}
	return bound, nil
}

func BindJob(
	scope platformtx.JobTenantScope,
	effects Effects,
	actorID uuid.UUID,
) (usageapp.Store, error) {
	if effects == nil || actorID == uuid.Nil || actorID.Version() != uuid.Version(7) {
		return nil, usageapp.ErrConfiguration
	}
	binder, err := platformtx.NewJobOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				tenantID: scope.TenantID(), actorID: actorID,
				queryRow: func(ctx context.Context, sql string, arguments ...any) row {
					return database.QueryRow(ctx, sql, arguments...)
				},
				execute: func(ctx context.Context, sql string, arguments ...any) error {
					_, executeErr := database.Execute(ctx, sql, arguments...)
					return executeErr
				},
				notFound: platformpostgres.RowNotFound,
				effects:  effects,
				allowed: func(ctx context.Context) bool {
					return platformtx.ContextOwnsJobScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx)
				},
			}
		},
	)
	if err != nil {
		return nil, usageapp.ErrConfiguration
	}
	bound, err := platformtx.BindJob(scope, binder)
	if err != nil || bound == nil {
		return nil, usageapp.ErrConfiguration
	}
	return bound, nil
}

func BindQueries(scope platformtx.TenantScope) (usageapp.CurrentUsageStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *queryStore {
			return &queryStore{
				tenantID: scope.TenantID(),
				queryRow: func(ctx context.Context, statement string, arguments ...any) row {
					return database.QueryRow(ctx, statement, arguments...)
				},
				query: func(
					ctx context.Context, statement string, arguments ...any,
				) (queryRows, error) {
					return database.Query(ctx, statement, arguments...)
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
		return nil, usageapp.ErrQueryConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, usageapp.ErrQueryConfiguration
	}
	return bound, nil
}
