package rolepostgres

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
func scanRole(row scanner) (identitydomain.Role, error) {
	var result identitydomain.Role
	var id, scopeID uuid.UUID
	var organizationID *uuid.UUID
	var retiredAt *time.Time
	if row.Scan(&id, &result.ScopeKind, &scopeID, &organizationID, &result.Name, &result.Description, &result.Status, &result.Protected, &result.Default, &result.Revision, &result.CreatedAt, &result.UpdatedAt, &retiredAt) != nil {
		return identitydomain.Role{}, identitydomain.ErrStorage
	}
	result.ID = identitydomain.IdentityID(id)
	result.ScopeID = identitydomain.IdentityID(scopeID)
	if organizationID != nil {
		result.OrganizationID = identitydomain.IdentityID(*organizationID)
	}
	if retiredAt != nil {
		result.RetiredAt = *retiredAt
	}
	return result, nil
}
