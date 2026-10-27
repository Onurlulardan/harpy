package platformiampostgres

import (
	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	"github.com/google/uuid"
)

func scanMutation(row scanner) (platformdomain.MutationResult, error) {
	var target, subject, role, subjectRole, rolePermission uuid.UUID
	var result platformdomain.MutationResult
	err := row.Scan(&target, &subject, &role, &subjectRole, &rolePermission,
		&result.TargetAuthorizationVersion, &result.PlatformClock, &result.Replayed)
	if err != nil {
		return platformdomain.MutationResult{}, storeError(err)
	}
	result.TargetUserID = domainID(target)
	result.SubjectID = domainID(subject)
	result.RoleID = domainID(role)
	result.SubjectRoleBindingID = domainID(subjectRole)
	result.RolePermissionBindingID = domainID(rolePermission)
	return result, nil
}
