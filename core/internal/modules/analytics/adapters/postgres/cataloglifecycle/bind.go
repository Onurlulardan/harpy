package cataloglifecycle

import (
	"context"
	_ "embed"

	catalogownerapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/catalogcomponentowner"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const owner = "analytics"

//go:embed sql/tenant/binding.sql
var bindingSQL string

//go:embed sql/tenant/preflight.sql
var preflightSQL string

//go:embed sql/tenant/record_purge_authority.sql
var recordPurgeAuthoritySQL string

//go:embed sql/tenant/read_purge_authority.sql
var readPurgeAuthoritySQL string

type Binder struct{}

func NewBinder() *Binder { return &Binder{} }

func (binder *Binder) Bind(
	ctx context.Context,
	scope platformtx.TenantScope,
	generation int64,
) (catalogownerapp.LifecycleStore, error) {
	if binder == nil || generation < 1 || !scopeUsable(ctx, scope) {
		return nil, platformpostgres.ErrOwnerBindingUnauthorized
	}
	authority, err := platformtx.MintOwnerAuthority(scope, owner)
	if err != nil || !authority.OwnsScope(scope) ||
		authority.Generation() != generation {
		return nil, platformpostgres.ErrOwnerBindingUnauthorized
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(
		owner,
		func(database platformpostgres.BoundExecutor) *Store {
			return &Store{
				generation: generation,
				queryRow: func(
					callCtx context.Context,
					statement string,
					arguments ...any,
				) Row {
					return database.QueryRow(
						callCtx,
						statement,
						arguments...,
					)
				},
				allowed: func(callCtx context.Context) bool {
					return scopeUsable(callCtx, scope)
				},
			}
		},
	)
	if err != nil {
		return nil, platformpostgres.ErrOwnerBindingUnauthorized
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, platformpostgres.ErrOwnerBindingUnauthorized
	}
	return bound, nil
}

func scopeUsable(
	ctx context.Context,
	scope platformtx.TenantScope,
) bool {
	return ctx != nil && ctx.Err() == nil &&
		!platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireOwner(scope, owner) == nil
}

var _ catalogownerapp.LifecycleStoreBinder = (*Binder)(nil)
