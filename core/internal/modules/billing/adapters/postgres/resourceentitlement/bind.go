package resourceentitlement

import (
	"context"

	resourceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/resourceentitlement"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func Bind(scope platformtx.TenantScope) (resourceapp.ResourceStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				tenantID: scope.TenantID(), actorID: scope.ActorID(),
				queryRow: func(ctx context.Context, statement string, arguments ...any) row {
					return database.QueryRow(ctx, statement, arguments...)
				},
				execute: func(ctx context.Context, statement string, arguments ...any) error {
					_, executeErr := database.Execute(ctx, statement, arguments...)
					return executeErr
				},
				allowed: func(ctx context.Context) bool {
					return platformtx.ContextOwnsScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx) &&
						platformtx.RequireActive(scope, platformtx.Command) == nil
				},
			}
		},
	)
	if err != nil {
		return nil, resourceapp.ErrConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, resourceapp.ErrConfiguration
	}
	return bound, nil
}
