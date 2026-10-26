package profilepostgres

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	identityports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/identity"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type tenantScope = platformtx.TenantScope
type scopeMode = platformtx.ScopeMode

const (
	commandMode  scopeMode = platformtx.Command
	snapshotMode scopeMode = platformtx.SnapshotQuery
)

type boundStore struct {
	database platformpostgres.BoundExecutor
}
type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (identityports.ProfileAliases[tenantScope], error) {
	return newStore()
}

func newStore() (*store, error) {
	binder, err := platformpostgres.NewOwnerBinder("iam", func(database platformpostgres.BoundExecutor) *boundStore {
		return &boundStore{database: database}
	})
	if err != nil {
		return nil, identitydomain.ErrStorage
	}
	return &store{binder: binder, valid: true}, nil
}

var _ identityports.ProfileAliases[tenantScope] = (*store)(nil)

func (storage *store) bind(scope tenantScope) (*boundStore, error) {
	if storage == nil || !storage.valid {
		return nil, identitydomain.ErrStorage
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, identitydomain.ErrStorage
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope tenantScope, mode scopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

func databaseID(id identitydomain.IdentityID) uuid.UUID { return uuid.UUID(id) }
