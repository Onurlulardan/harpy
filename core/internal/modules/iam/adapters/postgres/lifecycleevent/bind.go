package lifecycleevent

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantScope = platformtx.TenantScope
type credentialScope = platformtx.CredentialTenantScope

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type Sequences interface {
	Tenant(context.Context, platformtx.TenantScope, lifecycle.Change) (int64, error)
	Credential(context.Context, platformtx.CredentialTenantScope, lifecycle.Change) (int64, error)
}

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
}

func NewStore() (Sequences, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, lifecycle.ErrEffects
	}
	return &store{binder: binder}, nil
}

func (store *store) bindTenant(scope tenantScope) (*boundStore, error) {
	if store == nil {
		return nil, lifecycle.ErrEffects
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, lifecycle.ErrEffects
	}
	return bound, nil
}

func (store *store) bindCredential(scope credentialScope) (*boundStore, error) {
	if store == nil {
		return nil, lifecycle.ErrEffects
	}
	bound, err := platformtx.BindCredential(scope, store.binder)
	if err != nil || bound == nil {
		return nil, lifecycle.ErrEffects
	}
	return bound, nil
}

func tenantScopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope)
}

func credentialScopeUsable(ctx context.Context, scope credentialScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsCredentialScope(ctx, scope)
}
