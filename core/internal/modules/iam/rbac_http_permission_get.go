package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) GetPermission(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (RBACPermissionView, error) {
	if id == uuid.Nil {
		return RBACPermissionView{}, ErrRBACHTTPInvalid
	}
	var view RBACPermissionView
	err := application.inSnapshot(
		ctx, claims, "iam.permissions", "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			assignment, readErr := application.runtime.assignments.Get(
				scoped, scope, rbacdomain.AssignmentGetQuery{
					ID: rbacdomain.ID(id), Evidence: rbacQueryEvidence(claims, permissionID),
				},
			)
			if readErr != nil {
				return readErr
			}
			view, readErr = application.assignmentView(
				scoped, scope, claims, permissionID, assignment,
			)
			return readErr
		},
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	if !validRBACPermissionView(view) {
		return RBACPermissionView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
