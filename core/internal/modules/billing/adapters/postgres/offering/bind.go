package offering

import (
	"context"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func Bind(scope platformtx.GlobalScope) (offeringapp.Store, error) {
	binder, err := platformpostgres.NewGlobalOwnerBinder(
		"billing",
		func(database platformpostgres.BoundExecutor) *store {
			return &store{
				queryRow: func(ctx context.Context, sql string, arguments ...any) row {
					return database.QueryRow(ctx, sql, arguments...)
				},
				execute: func(ctx context.Context, sql string, arguments ...any) error {
					_, err := database.Execute(ctx, sql, arguments...)
					return err
				},
				allowed: func(ctx context.Context) bool {
					return platformtx.ContextOwnsGlobalScope(ctx, scope) &&
						!platformtx.NetworkAllowed(ctx) && scope.Mode() == platformtx.Command
				},
			}
		},
	)
	if err != nil {
		return nil, offeringapp.ErrConfiguration
	}
	bound, err := platformtx.BindGlobal(scope, binder)
	if err != nil || bound == nil {
		return nil, offeringapp.ErrConfiguration
	}
	return bound, nil
}
