package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

func (store *invitationStore) Revoke(
	ctx context.Context,
	scope tenantScope,
	command invitation.RevokeCommand,
) (invitation.Ref, error) {
	if command.Validate() != nil || !invitationCommandScopeUsable(ctx, scope) ||
		scope.ActorID() != command.ActorID {
		return invitation.Ref{}, invitation.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return invitation.Ref{}, err
	}
	return scanInvitationReference(bound.database.QueryRow(
		ctx, invitationRevokeSQL, command.OperationID, command.RequestSHA256[:],
		command.ID, command.ExpectedRevision,
		command.ActorID, command.ExpectedAuthorizationVersion,
	))
}
