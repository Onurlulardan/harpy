package userrelationpostgres

import (
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func scanMutation(row commandRow) (identitydomain.MutationResult, error) {
	var result identitydomain.MutationResult
	var id uuid.UUID
	if row == nil || row.Scan(
		&id, &result.Revision, &result.AuthorizationClock, &result.Replayed,
	) != nil {
		return identitydomain.MutationResult{}, identitydomain.ErrStorage
	}
	result.ID = identitydomain.IdentityID(id)
	if result.Validate() != nil {
		return identitydomain.MutationResult{}, identitydomain.ErrStorage
	}
	return result, nil
}
