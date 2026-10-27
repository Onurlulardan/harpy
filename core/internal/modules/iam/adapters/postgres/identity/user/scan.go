package userpostgres

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

func scanUser(row scanner) (identitydomain.User, error) {
	var result identitydomain.User
	var id, identifierID uuid.UUID
	var roleIDs, organizationIDs []uuid.UUID
	var accessModes []string
	var tombstonedAt *time.Time
	if row.Scan(
		&id, &result.Status, &result.TokenVersion,
		&result.AuthorizationVersion, &result.Revision, &identifierID,
		&result.EmailDisplay, &result.EmailNormalized, &roleIDs,
		&organizationIDs, &result.CreatedAt, &result.UpdatedAt, &tombstonedAt,
		&result.EmailVerified, &accessModes,
	) != nil {
		return identitydomain.User{}, identitydomain.ErrStorage
	}
	result.ID = identitydomain.IdentityID(id)
	result.IdentifierID = identitydomain.IdentityID(identifierID)
	result.RoleIDs = identityIDs(roleIDs)
	result.OrganizationIDs = identityIDs(organizationIDs)
	result.AccessModes = append([]string(nil), accessModes...)
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
