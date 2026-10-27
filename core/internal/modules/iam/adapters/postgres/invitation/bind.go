package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type invitationDatabase interface {
	Execute(context.Context, string, ...any) (platformpostgres.CommandResult, error)
	Query(context.Context, string, ...any) (platformpostgres.Rows, error)
	QueryRow(context.Context, string, ...any) platformpostgres.Row
}

type tenantScope = platformtx.TenantScope
type credentialScope = platformtx.CredentialTenantScope
type invitationRow = platformpostgres.Row

type boundInvitationStore struct{ database invitationDatabase }

type invitationStore struct {
	binder platformpostgres.OwnerBinder[*boundInvitationStore]
	valid  bool
}

func NewInvitationStore() (invitation.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam",
		func(database platformpostgres.BoundExecutor) *boundInvitationStore {
			return &boundInvitationStore{database: database}
		},
	)
	if err != nil {
		return nil, invitation.ErrUnavailable
	}
	return &invitationStore{binder: binder, valid: true}, nil
}

var _ invitation.Store = (*invitationStore)(nil)

func (store *invitationStore) bindTenant(
	scope tenantScope,
) (*boundInvitationStore, error) {
	if store == nil || !store.valid {
		return nil, invitation.ErrUnavailable
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, invitation.ErrUnavailable
	}
	return bound, nil
}

func (store *invitationStore) bindCredential(
	scope credentialScope,
) (*boundInvitationStore, error) {
	if store == nil || !store.valid {
		return nil, invitation.ErrUnavailable
	}
	bound, err := platformtx.BindCredential(scope, store.binder)
	if err != nil || bound == nil {
		return nil, invitation.ErrUnavailable
	}
	return bound, nil
}

func invitationCommandScopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}

func invitationSnapshotScopeUsable(ctx context.Context, scope tenantScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) &&
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) == nil
}

func invitationCredentialScopeUsable(ctx context.Context, scope credentialScope) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsCredentialScope(ctx, scope) &&
		platformtx.RequireCredentialPurpose(
			scope, platformtx.CredentialInvitation,
		) == nil
}
