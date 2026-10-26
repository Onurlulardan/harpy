package discount

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func Bind(scope platformtx.TenantScope) (discountapp.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				tenantID: scope.TenantID(),
				queryRow: func(ctx context.Context, sql string, arguments ...any) row {
					return database.QueryRow(ctx, sql, arguments...)
				},
				execute: func(ctx context.Context, sql string, arguments ...any) error {
					_, err := database.Execute(ctx, sql, arguments...)
					return err
				},
				notFound: platformpostgres.RowNotFound,
				allowed: func(ctx context.Context) bool {
					return platformtx.ContextOwnsScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx) &&
						platformtx.RequireActive(scope, platformtx.Command) == nil
				},
			}
		},
	)
	if err != nil {
		return nil, discountapp.ErrConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, discountapp.ErrConfiguration
	}
	return bound, nil
}
