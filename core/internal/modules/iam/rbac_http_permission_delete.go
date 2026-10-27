package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) DeletePermission(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACMutation,
) error {
	if !rbacMutationValid(input) {
		return ErrRBACHTTPInvalid
	}
	const command = "iam.rbac.assignment.retire"
	payload := struct {
		ID               uuid.UUID
		ExpectedRevision int64
	}{input.ID, input.ExpectedRevision}
	operationID, requestSHA, stepUp, err := application.mutationContext(
		claims, command, input.ID, input.IdempotencyKey, payload,
	)
	if err != nil {
		return err
	}
	return application.inCommand(
		ctx, claims, operationID.String(), "delete",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, retireErr := application.runtime.assignments.Retire(
				scoped, scope, rbacdomain.AssignmentMutation{
					ID: rbacdomain.ID(input.ID), ExpectedRevision: input.ExpectedRevision,
					StepUp:   stepUp,
					Evidence: rbacMutationEvidence(claims, permissionID, operationID, requestSHA),
				},
			)
			return retireErr
		},
	)
}
