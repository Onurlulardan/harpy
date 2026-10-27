package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

func (store *invitationStore) Read(
	ctx context.Context,
	scope tenantScope,
	query invitation.ReadQuery,
) (invitation.Invitation, error) {
	if query.Validate() != nil || !invitationSnapshotScopeUsable(ctx, scope) {
		return invitation.Invitation{}, invitation.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return invitation.Invitation{}, err
	}
	return scanInvitation(bound.database.QueryRow(
		ctx, invitationReadSQL, scope.ActorID(), query.ID,
	))
}
