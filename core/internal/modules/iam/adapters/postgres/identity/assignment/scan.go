package assignmentpostgres

import (
	"time"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

type scanner interface{ Scan(...any) error }

func scanMutation(row commandRow) (identitydomain.MutationResult, error) {
	var result identitydomain.MutationResult
	var id uuid.UUID
	if row == nil || row.Scan(&id, &result.Revision, &result.AuthorizationClock, &result.Replayed) != nil {
		return identitydomain.MutationResult{}, identitydomain.ErrStorage
	}
	result.ID = identitydomain.IdentityID(id)
	if result.Validate() != nil {
		return identitydomain.MutationResult{}, identitydomain.ErrStorage
	}
	return result, nil
}

func scanAssignment(row scanner) (identitydomain.Assignment, error) {
	var result identitydomain.Assignment
	var id, userID, roleID, scopeID uuid.UUID
	var organizationID, membershipID *uuid.UUID
	var revokedAt *time.Time
	if row == nil || row.Scan(
		&id, &userID, &roleID, &result.ScopeKind, &scopeID, &organizationID,
		&membershipID, &result.State, &result.Revision, &result.CreatedAt,
		&result.UpdatedAt, &revokedAt,
	) != nil {
		return identitydomain.Assignment{}, identitydomain.ErrStorage
	}
	result.ID = identitydomain.IdentityID(id)
	result.UserID = identitydomain.IdentityID(userID)
	result.RoleID = identitydomain.IdentityID(roleID)
	result.ScopeID = identitydomain.IdentityID(scopeID)
	if organizationID != nil {
		result.OrganizationID = identitydomain.IdentityID(*organizationID)
	}
	if membershipID != nil {
		result.MembershipID = identitydomain.IdentityID(*membershipID)
	}
	if revokedAt != nil {
		result.RevokedAt = *revokedAt
	}
	if result.Validate() != nil {
		return identitydomain.Assignment{}, identitydomain.ErrStorage
	}
	return result, nil
}
