package postgres

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	"github.com/google/uuid"
)

func scanInvitation(row invitationRow) (invitation.Invitation, error) {
	var result invitation.Invitation
	var scopeKind, state string
	var organizationID, acceptedUserID, acceptedIdentifierID *uuid.UUID
	var terminalAt *time.Time
	err := row.Scan(
		&result.ID, &result.DisplayValue, &result.NormalizedValue,
		&result.RoleID, &result.RoleName, &scopeKind, &result.RoleScope.ID,
		&organizationID, &result.InvitedBy, &acceptedUserID,
		&acceptedIdentifierID, &state, &result.Revision,
		&result.CreatedAt, &result.UpdatedAt, &result.ExpiresAt, &terminalAt,
	)
	if err != nil {
		return invitation.Invitation{}, mapInvitationError(err)
	}
	result.RoleScope.Kind = invitation.RoleScopeKind(scopeKind)
	result.State = invitation.State(state)
	if organizationID != nil {
		result.RoleScope.OrganizationID = *organizationID
	}
	if acceptedUserID != nil {
		result.AcceptedUserID = *acceptedUserID
	}
	if acceptedIdentifierID != nil {
		result.AcceptedIdentifierID = *acceptedIdentifierID
	}
	if terminalAt != nil {
		result.TerminalAt = *terminalAt
	}
	if result.Validate() != nil {
		return invitation.Invitation{}, invitation.ErrUnavailable
	}
	return result, nil
}
