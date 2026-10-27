package legalhold

import (
	"context"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func Bind(scope platformtx.TenantScope) (legalholdapp.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				tenantID: scope.TenantID(),
				database: DatabaseAccess{
					QueryRow: func(
						ctx context.Context, statement string, arguments ...any,
					) Row {
						return database.QueryRow(ctx, statement, arguments...)
					},
					Allowed: func(ctx context.Context) bool {
						return platformtx.ContextOwnsScope(ctx, scope) &&
							!platformtx.NetworkAllowed(ctx) &&
							platformtx.RequireActive(scope, platformtx.Command) == nil
					},
				},
			}
		},
	)
	if err != nil {
		return nil, legalholdapp.ErrConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, legalholdapp.ErrConfiguration
	}
	return bound, nil
}
