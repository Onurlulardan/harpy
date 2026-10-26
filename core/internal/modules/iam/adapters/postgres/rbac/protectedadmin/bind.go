package protectedadminpostgres

import (
	"context"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type tenantScope = platformtx.TenantScope

const (
	commandMode  = platformtx.Command
	snapshotMode = platformtx.SnapshotQuery
)

type scanner interface{ Scan(...any) error }
type boundStore struct {
	database platformpostgres.BoundExecutor
}
type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (rbacports.ProtectedAdminStore[tenantScope], error) {
	binder, err := platformpostgres.NewOwnerBinder("iam", func(database platformpostgres.BoundExecutor) *boundStore { return &boundStore{database: database} })
	if err != nil {
		return nil, rbacdomain.ErrStorage
	}
	return &store{binder: binder, valid: true}, nil
}
func (storage *store) bind(scope tenantScope) (*boundStore, error) {
	if storage == nil || !storage.valid {
		return nil, rbacdomain.ErrStorage
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, rbacdomain.ErrStorage
	}
	return bound, nil
}
func scopeUsable(ctx context.Context, scope tenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) && platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}
func databaseID(id rbacdomain.ID) uuid.UUID { return uuid.UUID(id) }

var _ rbacports.ProtectedAdminStore[tenantScope] = (*store)(nil)
