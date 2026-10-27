package membershippostgres

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

func scanMembership(row scanner) (identitydomain.Membership, error) {
	var result identitydomain.Membership
	var id, organizationID, userID uuid.UUID
	var tombstonedAt *time.Time
	if row == nil || row.Scan(
		&id, &organizationID, &userID, &result.Kind, &result.State,
		&result.Revision, &result.CreatedAt, &result.UpdatedAt, &tombstonedAt,
	) != nil {
		return identitydomain.Membership{}, identitydomain.ErrStorage
	}
	result.ID = identitydomain.IdentityID(id)
	result.OrganizationID = identitydomain.IdentityID(organizationID)
	result.UserID = identitydomain.IdentityID(userID)
	if tombstonedAt != nil {
		result.TombstonedAt = *tombstonedAt
	}
	if result.Validate() != nil {
		return identitydomain.Membership{}, identitydomain.ErrStorage
	}
	return result, nil
}
