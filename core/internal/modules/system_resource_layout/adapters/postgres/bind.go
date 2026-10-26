package postgres

import (
	"context"

	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func BindStore(scope platformtx.TenantScope) (layoutapp.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		storageOwner,
		func(database platformpostgres.BoundExecutor) *layoutStore {
			return &layoutStore{
				queryRow: func(
					ctx context.Context,
					query string,
					arguments ...any,
				) row {
					return database.QueryRow(ctx, query, arguments...)
				},
				commandAllowed: func(ctx context.Context) bool {
					return scopeUsable(ctx, scope, platformtx.Command)
				},
				queryAllowed: func(ctx context.Context) bool {
					return scopeUsable(ctx, scope, platformtx.Command) ||
						scopeUsable(ctx, scope, platformtx.SnapshotQuery)
				},
			}
		},
	)
	if err != nil {
		return nil, layoutUnavailable()
	}
	store, err := platformtx.Bind(scope, binder)
	if err != nil || store == nil {
		return nil, layoutUnavailable()
	}
	return store, nil
}

func BindReferenceStore(
	scope platformtx.TenantScope,
) (layoutapp.ReferenceStore, error) {
	store, err := BindStore(scope)
	if err != nil {
		return nil, err
	}
	references, ok := store.(layoutapp.ReferenceStore)
	if !ok || references == nil {
		return nil, layoutUnavailable()
	}
	return references, nil
}

func scopeUsable(
	ctx context.Context,
	scope platformtx.TenantScope,
	mode platformtx.ScopeMode,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, mode) == nil
}
