package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

func (store *invitationStore) Create(
	ctx context.Context,
	scope tenantScope,
	command invitation.CreateCommand,
) (invitation.Ref, error) {
	if command.Validate() != nil || !invitationCommandScopeUsable(ctx, scope) ||
		command.Token.MatchesTenant(scope.TenantID()) != nil ||
		scope.ActorID() != command.InvitedBy {
		return invitation.Ref{}, invitation.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return invitation.Ref{}, err
	}
	return scanInvitationReference(bound.database.QueryRow(
		ctx, invitationCreateSQL,
		command.OperationID, command.RequestSHA256[:],
		command.ID, command.DisplayValue, command.NormalizedValue,
		command.NormalizedSHA256[:], command.RoleID,
		string(command.RoleScope.Kind), command.RoleScope.ID,
		nullableInvitationUUID(command.RoleScope.OrganizationID),
		command.InvitedBy, command.ExpectedAuthorizationVersion,
		command.Token.ID, command.Token.Selector[:], command.Token.HashKeyID,
		command.Token.SecretHMAC[:], command.Token.BindingSHA256[:],
		command.Token.DeliveryEnvelope.ArtifactID,
		command.Token.DeliveryArtifactRef, command.Token.DeliveryEnvelope.KeyID,
		command.Token.DeliveryEnvelope.Nonce[:],
		command.Token.DeliveryEnvelope.Ciphertext[:],
		command.Token.DeliveryEnvelope.Tag[:],
		command.Token.DeliveryEnvelope.AADSHA256[:],
		command.Token.DeliveryEnvelope.CiphertextSHA256[:],
	))
}
