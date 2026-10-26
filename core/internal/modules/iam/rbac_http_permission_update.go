package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) UpdatePermission(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACPermissionUpdate,
) (RBACPermissionView, error) {
	if !rbacPermissionUpdateValid(input) {
		return RBACPermissionView{}, ErrRBACHTTPInvalid
	}
	resource, actions, err := application.readPermissionCatalog(
		ctx, claims, input.ResourceID, input.ActionIDs,
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	const command = "iam.rbac.assignment.update"
	payload := struct {
		ID, ResourceID uuid.UUID
		Target         string
		TargetID       uuid.UUID
		ActionIDs      []uuid.UUID
		Revision       int64
	}{
		input.ID, input.ResourceID, string(input.Target), input.TargetID,
		identitySortedIDs(input.ActionIDs), input.ExpectedRevision,
	}
	operationID, requestSHA, stepUp, err := application.mutationContext(
		claims, command, input.ID, input.IdempotencyKey, payload,
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	var revision int64
	err = application.inCommand(
		ctx, claims, operationID.String(), "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			result, updateErr := application.runtime.assignments.Update(
				scoped, scope, rbacdomain.AssignmentMutation{
					ID: rbacdomain.ID(input.ID), ResourceID: rbacdomain.ID(input.ResourceID),
					Target: input.Target, TargetID: rbacdomain.ID(input.TargetID),
					Items:            rbacAssignmentItems(operationID, input.ActionIDs),
					ExpectedRevision: input.ExpectedRevision, StepUp: stepUp,
					Evidence: rbacMutationEvidence(claims, permissionID, operationID, requestSHA),
				},
			)
			revision = result.Revision
			return updateErr
		},
	)
	if err != nil {
		return RBACPermissionView{}, err
	}
	view := rbacPermissionMutationView(
		input.ID, resource, actions, input.Target, input.TargetID, revision,
	)
	if !validRBACPermissionView(view) {
		return RBACPermissionView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
