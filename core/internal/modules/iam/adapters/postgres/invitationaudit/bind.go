package invitationaudit

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var errUnavailable = errors.New("IAM invitation audit store is unavailable")

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type tenantScope = platformtx.TenantScope
type credentialScope = platformtx.CredentialTenantScope

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (invitation.AuditStore, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, errUnavailable
	}
	return &store{binder: binder, valid: true}, nil
}

func (store *store) bindTenant(
	scope tenantScope,
) (*boundStore, error) {
	if store == nil || !store.valid {
		return nil, errUnavailable
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, errUnavailable
	}
	return bound, nil
}

func (store *store) bindCredential(
	scope credentialScope,
) (*boundStore, error) {
	if store == nil || !store.valid {
		return nil, errUnavailable
	}
	bound, err := platformtx.BindCredential(scope, store.binder)
	if err != nil || bound == nil {
		return nil, errUnavailable
	}
	return bound, nil
}

func tenantUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}

func credentialUsable(
	ctx context.Context,
	scope credentialScope,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsCredentialScope(ctx, scope) &&
		platformtx.RequireCredentialPurpose(
			scope, platformtx.CredentialInvitation,
		) == nil
}
