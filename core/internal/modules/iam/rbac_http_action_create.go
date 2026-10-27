package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) CreateAction(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACCatalogCreate,
) (RBACActionView, error) {
	if !rbacCatalogCreateValid(input) {
		return RBACActionView{}, ErrRBACHTTPInvalid
	}
	const command = "iam.rbac.action.create"
	payload := struct{ Key, Name, Description string }{
		input.Key, input.DisplayName, input.Description,
	}
	operationID, requestSHA, _, err := application.mutationContext(
		claims, command, uuid.Nil, input.IdempotencyKey, payload,
	)
	if err != nil {
		return RBACActionView{}, err
	}
	id := identityChildID(operationID, "rbac-action")
	var revision int64
	err = application.inCommand(
		ctx, claims, operationID.String(), "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			result, createErr := application.runtime.actions.Create(
				scoped, scope, rbacdomain.ActionInput{
					ID: rbacdomain.ID(id), Key: input.Key,
					DisplayName: input.DisplayName, Description: input.Description,
					Evidence: rbacMutationEvidence(claims, permissionID, operationID, requestSHA),
				},
			)
			revision = result.Revision
			return createErr
		},
	)
	if err != nil {
		return RBACActionView{}, err
	}
	view := RBACActionView{
		ID: id.String(), Name: input.DisplayName, Slug: input.Key,
		Description: input.Description, State: "active", Revision: revision,
	}
	if !validRBACActionView(view) {
		return RBACActionView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
