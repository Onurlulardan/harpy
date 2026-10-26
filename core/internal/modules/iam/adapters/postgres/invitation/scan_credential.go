package postgres

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	"github.com/google/uuid"
)

func scanInvitationCredential(
	row invitationRow,
	result *invitation.CredentialRecord,
	selector, secretHMAC, binding *[]byte,
) error {
	var scopeKind, state string
	var organizationID, acceptedUserID, acceptedIdentifierID *uuid.UUID
	var terminalAt *time.Time
	err := row.Scan(
		&result.Invitation.ID, &result.Invitation.DisplayValue,
		&result.Invitation.NormalizedValue, &result.Invitation.RoleID,
		&result.Invitation.RoleName, &scopeKind, &result.Invitation.RoleScope.ID,
		&organizationID, &result.Invitation.InvitedBy,
		&acceptedUserID, &acceptedIdentifierID,
		&state, &result.Invitation.Revision, &result.Invitation.CreatedAt,
		&result.Invitation.UpdatedAt, &result.Invitation.ExpiresAt, &terminalAt,
		&result.Proof.ID, selector, &result.Proof.HashKeyID,
		secretHMAC, binding, &result.Proof.Revision, &result.ProfileCatalog,
	)
	if err != nil {
		return mapInvitationError(err)
	}
	result.Invitation.RoleScope.Kind = invitation.RoleScopeKind(scopeKind)
	result.Invitation.State = invitation.State(state)
	if organizationID != nil {
		result.Invitation.RoleScope.OrganizationID = *organizationID
	}
	if acceptedUserID != nil {
		result.Invitation.AcceptedUserID = *acceptedUserID
	}
	if acceptedIdentifierID != nil {
		result.Invitation.AcceptedIdentifierID = *acceptedIdentifierID
	}
	if terminalAt != nil {
		result.Invitation.TerminalAt = *terminalAt
	}
	return nil
}
