package permissioncachepostgres

import (
	"context"

	permissioncacheapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissioncache"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type boundStore struct {
	database platformpostgres.BoundExecutor
}
type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
}

func New() (permissioncacheapp.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		})
	if err != nil {
		return nil, permissioncacheapp.ErrUnavailable
	}
	return &store{binder: binder}, nil
}

func (storage *store) bind(scope platformtx.TenantScope) (*boundStore, error) {
	if storage == nil {
		return nil, permissioncacheapp.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, permissioncacheapp.ErrUnavailable
	}
	return bound, nil
}

func usable(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}
