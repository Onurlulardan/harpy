package organizationpostgres

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

func scanOrganization(row scanner) (identitydomain.Organization, error) {
	var result identitydomain.Organization
	var id uuid.UUID
	var parentID *uuid.UUID
	var ownerIDs []uuid.UUID
	var tombstonedAt *time.Time
	if row.Scan(&id, &result.Name, &result.Slug, &result.Status, &parentID,
		&result.Revision, &result.CreatedAt, &result.UpdatedAt, &tombstonedAt,
		&ownerIDs) != nil {
		return identitydomain.Organization{}, identitydomain.ErrStorage
	}
	result.ID = identitydomain.IdentityID(id)
	result.OwnerIDs = identityIDs(ownerIDs)
	if parentID != nil {
		result.ParentID = identitydomain.IdentityID(*parentID)
	}
	if tombstonedAt != nil {
		result.TombstonedAt = *tombstonedAt
	}
	return result, nil
}

func identityIDs(values []uuid.UUID) []identitydomain.IdentityID {
	result := make([]identitydomain.IdentityID, len(values))
	for index, value := range values {
		result[index] = identitydomain.IdentityID(value)
	}
	return result
}
