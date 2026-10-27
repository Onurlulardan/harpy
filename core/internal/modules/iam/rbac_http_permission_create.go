package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) CreatePermission(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACPermissionCreate,
) (RBACPermissionView, error) {
	if !rbacPermissionCreateValid(input) {
		return RBACPermissionView{}, ErrRBACHTTPInvalid
	}
	resource, actions, err := application.readPermissionCatalog(
		ctx, claims, input.ResourceID, input.ActionIDs,
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	const command = "iam.rbac.assignment.create"
	payload := struct {
		ResourceID uuid.UUID
		Target     string
		TargetID   uuid.UUID
		ActionIDs  []uuid.UUID
	}{input.ResourceID, string(input.Target), input.TargetID, identitySortedIDs(input.ActionIDs)}
	operationID, requestSHA, stepUp, err := application.mutationContext(
		claims, command, uuid.Nil, input.IdempotencyKey, payload,
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	id := identityChildID(operationID, "rbac-assignment")
	var revision int64
	err = application.inCommand(
		ctx, claims, operationID.String(), "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			result, createErr := application.runtime.assignments.Create(
				scoped, scope, rbacdomain.AssignmentMutation{
					ID: rbacdomain.ID(id), ResourceID: rbacdomain.ID(input.ResourceID),
					Target: input.Target, TargetID: rbacdomain.ID(input.TargetID),
					Items: rbacAssignmentItems(operationID, input.ActionIDs), StepUp: stepUp,
					Evidence: rbacMutationEvidence(claims, permissionID, operationID, requestSHA),
				},
			)
			revision = result.Revision
			return createErr
		},
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	view := rbacPermissionMutationView(
		id, resource, actions, input.Target, input.TargetID, revision,
	)
	if !validRBACPermissionView(view) {
		return RBACPermissionView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
