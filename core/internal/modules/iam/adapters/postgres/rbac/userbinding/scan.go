package userbindingpostgres

import (
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/google/uuid"
)

func scanMutation(row scanner) (rbacdomain.MutationResult, error) {
	var result rbacdomain.MutationResult
	var id uuid.UUID
	if row == nil || row.Scan(&id, &result.Revision, &result.AuthorizationClock, &result.Replayed) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	result.ID = rbacdomain.ID(id)
	if result.Validate(result.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return result, nil
}
func scanBinding(row scanner) (rbacdomain.Binding, error) {
	var result rbacdomain.Binding
	var id, subjectID, permissionID uuid.UUID
	if row == nil || row.Scan(&id, &subjectID, &permissionID, &result.State, &result.Revision) != nil {
		return rbacdomain.Binding{}, rbacdomain.ErrStorage
	}
	result.ID, result.SubjectID, result.PermissionID = rbacdomain.ID(id), rbacdomain.ID(subjectID), rbacdomain.ID(permissionID)
	return result, nil
}
