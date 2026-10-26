package protectedadminpostgres

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
func scanPolicy(row scanner) (rbacdomain.ProtectedAdminPolicy, error) {
	var result rbacdomain.ProtectedAdminPolicy
	var roleID, permissionID, bindingID uuid.UUID
	if row == nil || row.Scan(&roleID, &permissionID, &bindingID, &result.MinimumActiveAssignments, &result.AuthorizationEpoch, &result.Revision) != nil {
		return rbacdomain.ProtectedAdminPolicy{}, rbacdomain.ErrStorage
	}
	result.RoleID, result.TenantAllPermissionID, result.RolePermissionBindingID = rbacdomain.ID(roleID), rbacdomain.ID(permissionID), rbacdomain.ID(bindingID)
	return result, nil
}
