package invoice

import (
	"context"

	invoiceapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/invoice"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func Bind(scope platformtx.TenantScope, effects Effects) (invoiceapp.Store, error) {
	if effects == nil {
		return nil, invoiceapp.ErrConfiguration
	}
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
		return nil, invoiceapp.ErrConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, invoiceapp.ErrConfiguration
	}
	return bound, nil
}
