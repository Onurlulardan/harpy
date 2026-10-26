package profile

import (
	"context"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = platformtx.TenantScope
type commandRow = platformpostgres.Row

type boundCommandStore struct {
	database platformpostgres.BoundExecutor
}

type commandStore struct {
	binder platformpostgres.OwnerBinder[*boundCommandStore]
	valid  bool
}

func NewCommandStore() (profilecommand.CommandPort, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundCommandStore {
			return &boundCommandStore{database: database}
		},
	)
	if err != nil {
		return nil, profilecommand.ErrEvidenceMissing
	}
	return &commandStore{binder: binder, valid: true}, nil
}

var _ profilecommand.CommandPort = (*commandStore)(nil)

func (store *commandStore) bind(scope tenantScope) (*boundCommandStore, error) {
	if store == nil || !store.valid {
		return nil, profilecommand.ErrEvidenceMissing
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, profilecommand.ErrEvidenceMissing
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}

func snapshotScopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
}

func readScopeUsable(ctx context.Context, scope tenantScope) bool {
	return snapshotScopeUsable(ctx, scope) || scopeUsable(ctx, scope)
}
