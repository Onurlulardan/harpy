package permissionpostgres

import (
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/google/uuid"
)

func scanMutation(row scanner) (rbacdomain.MutationResult, error) {
	var result rbacdomain.MutationResult
	var id uuid.UUID
	if row == nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	if err := row.Scan(&id, &result.Revision, &result.AuthorizationClock, &result.Replayed); err != nil {
		return rbacdomain.MutationResult{}, mapError(err, rbacdomain.ErrInvalidMutation)
	}
	result.ID = rbacdomain.ID(id)
	if result.Validate(result.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return result, nil
}
func scanPermission(row scanner) (rbacdomain.Permission, error) {
	var result rbacdomain.Permission
	var id uuid.UUID
	var resourceID, actionID *uuid.UUID
	if row == nil {
		return rbacdomain.Permission{}, rbacdomain.ErrStorage
	}
	if err := row.Scan(&id, &result.MatchKind, &resourceID, &actionID, &result.Protected,
		&result.State, &result.Revision); err != nil {
		return rbacdomain.Permission{}, mapError(err, rbacdomain.ErrInvalidQuery)
	}
	result.ID = rbacdomain.ID(id)
	if resourceID != nil {
		result.ResourceID = rbacdomain.ID(*resourceID)
	}
	if actionID != nil {
		result.ActionID = rbacdomain.ID(*actionID)
	}
	return result, nil
}
