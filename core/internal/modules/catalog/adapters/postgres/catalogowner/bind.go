package catalogowner

import (
	"context"
	_ "embed"

	catalogcomponent "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogcomponent/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/demohandoff"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/demomaterialization"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const owner = "catalog"

//go:embed sql/tenant/read.sql
var readSQL string

//go:embed sql/tenant/apply.sql
var applySQL string

//go:embed sql/tenant/blockers.sql
var blockersSQL string

//go:embed sql/tenant/allows.sql
var allowsSQL string

type binder struct {
	factory *catalogcomponent.StoreFactory
}

func (value *binder) Bind(
	ctx context.Context, scope platformtx.TenantScope,
) (catalogcomponent.Store, error) {
	if value == nil || value.factory == nil || !scopeUsable(ctx, scope) {
		return nil, platformpostgres.ErrOwnerBindingUnauthorized
	}
	authority, err := platformtx.MintOwnerAuthority(scope, owner)
	if err != nil || !authority.OwnsScope(scope) || authority.Generation() < 1 {
		return nil, platformpostgres.ErrOwnerBindingUnauthorized
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(owner,
		func(database platformpostgres.BoundExecutor) *catalogcomponent.BoundStore {
			store, buildErr := value.factory.Build(authority.Generation(),
				func(callCtx context.Context, statement string, arguments ...any) catalogcomponent.Row {
					return database.QueryRow(callCtx, statement, arguments...)
				}, func(callCtx context.Context) bool { return scopeUsable(callCtx, scope) })
			if buildErr != nil {
				return nil
			}
			return store
		})
	if err != nil {
		return nil, err
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, platformpostgres.ErrOwnerBindingUnauthorized
	}
	return bound, nil
}

func scopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && platformtx.RequireOwner(scope, owner) == nil
}

type provisionBinder struct{}

func (binder *provisionBinder) Bind(
	ctx context.Context,
	scope platformtx.TenantScope,
) (demomaterialization.Store, error) {
	if binder == nil || !provisionScopeUsable(ctx, scope) {
		return nil, demomaterialization.ErrUnavailable
	}
	authority, err := platformtx.MintOwnerAuthority(scope, owner)
	if err != nil || authority.Generation() < 1 {
		return nil, demomaterialization.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(owner,
		func(database platformpostgres.BoundExecutor) *demomaterialization.BoundStore {
			store, buildErr := demomaterialization.NewBoundStore(
				authority.Generation(), provisionSQL,
				func(callCtx context.Context, statement string,
					arguments ...any,
				) demomaterialization.Row {
					return database.QueryRow(callCtx, statement, arguments...)
				},
				func(callCtx context.Context) bool {
					return provisionScopeUsable(callCtx, scope)
				},
			)
			if buildErr != nil {
				return nil
			}
			return store
		})
	if err != nil {
		return nil, demomaterialization.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, demomaterialization.ErrUnavailable
	}
	return bound, nil
}

func provisionScopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	return scopeUsable(ctx, scope) && scope.Mode() == platformtx.Command &&
		scope.OperationID() == demomaterialization.ProvisionOperation
}

func NewDemoHandoffBinder() demohandoff.StoreBinder {
	return func(
		ctx context.Context,
		scope platformtx.TenantScope,
		authority platformtx.OwnerAuthority,
	) (demohandoff.Store, error) {
		if !provisionScopeUsable(ctx, scope) ||
			!authority.OwnsScope(scope) || authority.Domain() != owner {
			return nil, platformpostgres.ErrOwnerBindingUnauthorized
		}
		ownerBinder, err := platformpostgres.NewOwnerBinder(owner,
			func(database platformpostgres.BoundExecutor) *demohandoff.BoundStore {
				store, buildErr := demohandoff.NewBoundStore(
					demoHandoffSQL,
					func(callCtx context.Context, statement string,
						arguments ...any,
					) demohandoff.Row {
						return database.QueryRow(callCtx, statement, arguments...)
					},
					func(callCtx context.Context) bool {
						return provisionScopeUsable(callCtx, scope)
					},
				)
				if buildErr != nil {
					return nil
				}
				return store
			})
		if err != nil {
			return nil, err
		}
		bound, err := platformtx.Bind(scope, ownerBinder)
		if err != nil || bound == nil {
			return nil, platformpostgres.ErrOwnerBindingUnauthorized
		}
		return bound, nil
	}
}

var _ demomaterialization.StoreBinder = (*provisionBinder)(nil)
