package catalogquery

import (
	"context"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func Bind(scope platformtx.GlobalScope) (catalogapp.Store, error) {
	binder, err := platformpostgres.NewGlobalOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				query: func(ctx context.Context, statement string, arguments ...any) (rows, error) {
					return database.Query(ctx, statement, arguments...)
				},
				allowed: func(ctx context.Context) bool {
					return platformtx.ContextOwnsGlobalScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx) &&
						scope.Mode() == platformtx.SnapshotQuery
				},
			}
		},
	)
	if err != nil {
		return nil, catalogapp.ErrConfiguration
	}
	bound, err := platformtx.BindGlobal(scope, binder)
	if err != nil || bound == nil {
		return nil, catalogapp.ErrConfiguration
	}
	return bound, nil
}
