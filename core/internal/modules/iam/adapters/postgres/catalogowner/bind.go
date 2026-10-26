package catalogowner

import (
	"context"
	_ "embed"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	catalogcomponent "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogcomponent/api"
	catalogprojection "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogprojection/api"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const owner = "iam"

//go:embed sql/tenant/read.sql
var readSQL string

//go:embed sql/tenant/apply.sql
var applySQL string

//go:embed sql/tenant/blockers.sql
var blockersSQL string

//go:embed sql/tenant/allows.sql
var allowsSQL string

type Owner struct {
	*catalogcomponent.PersistentOwner
	*catalogprojection.Application
}

func New() (*Owner, error) {
	factory, err := catalogcomponent.NewStoreFactory(owner, catalogcomponent.Queries{
		Read: readSQL, Apply: applySQL, Blockers: blockersSQL, Allows: allowsSQL,
	})
	if err != nil {
		return nil, err
	}
	persistent, err := catalogcomponent.NewPersistentOwner(owner, &binder{factory: factory})
	if err != nil {
		return nil, err
	}
	provision, err := catalogprojection.New(
		owner, catalogpackagev1.ContractIAM,
		projectionDecoder{}, NewProvisionBinder(),
	)
	if err != nil {
		return nil, err
	}
	return &Owner{PersistentOwner: persistent, Application: provision}, nil
}

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

type ProvisionBinder struct{}

func NewProvisionBinder() *ProvisionBinder { return &ProvisionBinder{} }

func (binder *ProvisionBinder) Bind(
	ctx context.Context,
	scope platformtx.TenantScope,
) (catalogprojection.Store, error) {
	if binder == nil || !provisionScopeUsable(ctx, scope) {
		return nil, catalogprojection.ErrUnavailable
	}
	authority, err := platformtx.MintOwnerAuthority(scope, owner)
	if err != nil || authority.Generation() < 1 {
		return nil, catalogprojection.ErrUnavailable
	}
	ownerBinder, err := platformpostgres.NewOwnerBinder(
		owner,
		func(database platformpostgres.BoundExecutor) *catalogprojection.BoundStore {
			store, buildErr := catalogprojection.NewBoundStore(
				owner, catalogpackagev1.ContractIAM, authority.Generation(),
				provisionSQL, []string{"IAM_TEMPLATE"},
				func(callCtx context.Context, sql string, arguments ...any) catalogprojection.Row {
					return database.QueryRow(callCtx, sql, arguments...)
				},
				func(callCtx context.Context) bool {
					return provisionScopeUsable(callCtx, scope)
				},
			)
			if buildErr != nil {
				return nil
			}
			return store
		},
	)
	if err != nil {
		return nil, catalogprojection.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, ownerBinder)
	if err != nil || bound == nil {
		return nil, catalogprojection.ErrUnavailable
	}
	return bound, nil
}

func provisionScopeUsable(ctx context.Context, scope platformtx.TenantScope) bool {
	return scopeUsable(ctx, scope) && scope.Mode() == platformtx.Command &&
		scope.OperationID() == catalogprojection.ProvisionOperation
}

var _ catalogprojection.StoreBinder = (*ProvisionBinder)(nil)
