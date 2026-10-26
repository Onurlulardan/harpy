package rbacchange

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type tenantScope = platformtx.TenantScope

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (rbacports.AuditStore[tenantScope], error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, rbacdomain.ErrStorage
	}
	return &store{binder: binder, valid: true}, nil
}

func (store *store) bind(scope tenantScope) (*boundStore, error) {
	if store == nil || !store.valid {
		return nil, rbacdomain.ErrStorage
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, rbacdomain.ErrStorage
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}

var _ rbacports.AuditStore[tenantScope] = (*store)(nil)
