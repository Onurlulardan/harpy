package provider

import (
	"context"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func BindCommand(
	scope platformtx.TenantScope,
	effects CommandEffects,
) (providerapp.CommandStore, error) {
	if effects == nil {
		return nil, providerapp.ErrConfiguration
	}
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *commandStore {
			return &commandStore{
				tenantID: scope.TenantID(), actorID: scope.ActorID(),
				database: databaseAccess{
					queryRow: func(
						ctx context.Context, statement string, arguments ...any,
					) databaseRow {
						return database.QueryRow(ctx, statement, arguments...)
					},
					allowed: func(ctx context.Context) bool {
						return platformtx.ContextOwnsScope(ctx, scope) &&
							!platformtx.NetworkAllowed(ctx) &&
							platformtx.RequireActive(scope, platformtx.Command) == nil
					},
				},
				effects: effects,
			}
		},
	)
	if err != nil {
		return nil, providerapp.ErrConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, providerapp.ErrConfiguration
	}
	return bound, nil
}

func NewBoundFinalizerStore(
	tenantID uuid.UUID,
	database platformtx.JobBoundExecutor,
) (providerapp.FinalizerStore, error) {
	if tenantID == uuid.Nil {
		return nil, providerapp.ErrConfiguration
	}
	return &finalizerStore{
		tenantID: tenantID,
		database: databaseAccess{
			queryRow: func(
				ctx context.Context, statement string, arguments ...any,
			) databaseRow {
				return database.QueryRow(ctx, statement, arguments...)
			},
			allowed: func(ctx context.Context) bool {
				return ctx != nil && ctx.Err() == nil &&
					!platformtx.NetworkAllowed(ctx)
			},
		},
	}, nil
}
