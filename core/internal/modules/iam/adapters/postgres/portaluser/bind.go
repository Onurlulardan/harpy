package portaluserpostgres

import (
	"context"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = platformtx.TenantScope
type scopeMode = platformtx.ScopeMode
type ownerAuthority = platformtx.OwnerAuthority

const (
	commandMode  = platformtx.Command
	snapshotMode = platformtx.SnapshotQuery
)

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (portaluser.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, portaluser.ErrUnavailable
	}
	return &store{binder: binder, valid: true}, nil
}

func (storage *store) bind(scope tenantScope) (*boundStore, error) {
	if storage == nil || !storage.valid {
		return nil, portaluser.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, storage.binder)
	if err != nil || bound == nil {
		return nil, portaluser.ErrUnavailable
	}
	return bound, nil
}

func usable(ctx context.Context, scope tenantScope, mode scopeMode) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireActive(scope, mode) == nil
}

func mintIAMOwnerAuthority(scope tenantScope) (ownerAuthority, error) {
	return platformtx.MintOwnerAuthority(scope, "iam")
}

var _ portaluser.Store = (*store)(nil)
