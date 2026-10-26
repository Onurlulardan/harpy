package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

func (store *invitationStore) Accept(
	ctx context.Context,
	scope credentialScope,
	command invitation.AcceptCommand,
) (invitation.AcceptResult, error) {
	if command.Validate() != nil || !invitationCredentialScopeUsable(ctx, scope) ||
		scope.LocatorBindingSHA256() != command.Token.BindingSHA256 {
		return invitation.AcceptResult{}, invitation.ErrInvalid
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return invitation.AcceptResult{}, err
	}
	var result invitation.AcceptResult
	err = bound.database.QueryRow(
		ctx, invitationAcceptSQL,
		command.OperationID, command.RequestSHA256[:], command.ID,
		command.ExpectedRevision, command.Token.ID, command.Token.Revision,
		command.Token.Selector[:], command.Token.HashKeyID,
		command.Token.SecretHMAC[:], command.Token.BindingSHA256[:],
		command.AcceptedUserID, command.AcceptedIdentifierID,
		command.AccessModeID, command.RoleAssignmentID,
		nullableInvitationUUID(command.OrganizationMembershipID),
		command.Profile.SchemaID, command.Profile.Values, command.Profile.SHA256[:],
		command.Password.Salt[:], command.Password.Tag[:],
		command.ExpectedAuthorizationVersion,
	).Scan(
		&result.Code, &result.Revision, &result.AuthorizationClock, &result.Replayed,
	)
	if err != nil {
		return invitation.AcceptResult{}, mapInvitationError(err)
	}
	var directoryRevision int64
	err = bound.database.QueryRow(
		ctx, invitationDirectoryUpsertSQL, scope.TenantID(),
		scope.AuthorityGroupGeneration(), scope.IAMGeneration(),
		command.AcceptedUserID, command.AcceptedIdentifierID,
		command.Directory.LookupKeyID, command.Directory.HMAC[:], int64(1),
		command.Directory.SourceSHA256[:],
	).Scan(&directoryRevision)
	if err != nil || directoryRevision != 1 {
		return invitation.AcceptResult{}, invitation.ErrUnavailable
	}
	if result.Code != "iam.invitation.accepted" || result.Revision < 2 ||
		result.AuthorizationClock < 1 {
		return invitation.AcceptResult{}, invitation.ErrUnavailable
	}
	return result, nil
}
