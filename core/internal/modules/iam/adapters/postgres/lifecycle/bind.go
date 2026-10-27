package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type lifecycleDatabase interface {
	Query(context.Context, string, ...any) (platformpostgres.Rows, error)
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type lifecycleRow interface{ Scan(...any) error }
type tenantScope = platformtx.TenantScope
type credentialScope = platformtx.CredentialTenantScope
type credentialPurpose = platformtx.CredentialPurpose

const (
	credentialEmailVerification = platformtx.CredentialEmailVerification
	credentialPasswordReset     = platformtx.CredentialPasswordReset
	credentialInitialAdminSetup = platformtx.CredentialInitialAdminSetup
	credentialPortalSetup       = platformtx.CredentialPortalSetup
)

func rowNotFound(err error) bool { return platformpostgres.RowNotFound(err) }

type boundStore struct{ database lifecycleDatabase }

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (lifecycle.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, lifecycle.ErrUnavailable
	}
	return &store{binder: binder, valid: true}, nil
}

var _ lifecycle.Store = (*store)(nil)

func (value *store) bindTenant(scope tenantScope) (*boundStore, error) {
	if value == nil || !value.valid {
		return nil, lifecycle.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, value.binder)
	if err != nil || bound == nil {
		return nil, lifecycle.ErrUnavailable
	}
	return bound, nil
}

func (value *store) bindCredential(
	scope credentialScope,
) (*boundStore, error) {
	if value == nil || !value.valid {
		return nil, lifecycle.ErrUnavailable
	}
	bound, err := platformtx.BindCredential(scope, value.binder)
	if err != nil || bound == nil {
		return nil, lifecycle.ErrUnavailable
	}
	return bound, nil
}

func commandUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}

func snapshotUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
}

func credentialUsable(
	ctx context.Context,
	scope credentialScope,
	purpose credentialPurpose,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsCredentialScope(ctx, scope) &&
		platformtx.RequireCredentialPurpose(scope, purpose) == nil
}
