package actionpostgres

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
func scanAction(row scanner) (rbacdomain.Action, error) {
	var result rbacdomain.Action
	var id uuid.UUID
	var description *string
	if row == nil {
		return rbacdomain.Action{}, rbacdomain.ErrStorage
	}
	if err := row.Scan(&id, &result.Key, &result.DisplayName, &description,
		&result.Protected, &result.State, &result.Revision); err != nil {
		return rbacdomain.Action{}, mapError(err, rbacdomain.ErrInvalidQuery)
	}
	result.ID = rbacdomain.ID(id)
	if description != nil {
		result.Description = *description
	}
	return result, nil
}
