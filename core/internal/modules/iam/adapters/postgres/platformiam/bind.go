package platformiampostgres

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/platformiam"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type tenantScope = platformtx.TenantScope
type scanner interface{ Scan(...any) error }

type boundStore struct {
	database platformpostgres.BoundExecutor
}
type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (platformports.Store[tenantScope], error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, platformdomain.ErrStorage
	}
	return &store{binder: binder, valid: true}, nil
}

func (storage *store) bind(scope tenantScope) (*boundStore, error) {
	if storage == nil || !storage.valid {
		return nil, platformdomain.ErrStorage
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, platformdomain.ErrStorage
	}
	return bound, nil
}

func commandScopeUsable(ctx context.Context, scope tenantScope) bool {
	return scopeUsable(ctx, scope, platformtx.Command)
}

func queryScopeUsable(ctx context.Context, scope tenantScope) bool {
	return scopeUsable(ctx, scope, platformtx.SnapshotQuery)
}

func scopeUsable(ctx context.Context, scope tenantScope, mode platformtx.ScopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

func databaseID(id platformdomain.ID) uuid.UUID { return uuid.UUID(id) }
func domainID(id uuid.UUID) platformdomain.ID   { return platformdomain.ID(id) }

var _ platformports.Store[tenantScope] = (*store)(nil)
var _ platformports.OwnerTransferStore[tenantScope] = (*store)(nil)
