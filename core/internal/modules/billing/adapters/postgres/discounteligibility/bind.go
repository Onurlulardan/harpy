package discounteligibility

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func Bind(scope platformtx.TenantScope) (discountapp.EligibilityStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				tenantID: scope.TenantID(),
				queryRow: func(ctx context.Context, statement string, arguments ...any) row {
					return database.QueryRow(ctx, statement, arguments...)
				},
				query: func(ctx context.Context, statement string, arguments ...any) (rows, error) {
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
		return nil, discountapp.ErrQueryConfiguration
	}
	bound, err := platformtx.Bind(scope, binder)
	if err != nil || bound == nil {
		return nil, discountapp.ErrQueryConfiguration
	}
	return bound, nil
}
