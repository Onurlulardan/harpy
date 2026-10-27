package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) UpdateAction(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACCatalogUpdate,
) (RBACActionView, error) {
	if !rbacCatalogUpdateValid(input) {
		return RBACActionView{}, ErrRBACHTTPInvalid
	}
	current, err := application.readAction(ctx, claims, input.ID)
	if err != nil {
		return RBACActionView{}, err
	}
	if input.Key != "" && input.Key != current.Key {
		return RBACActionView{}, ErrRBACHTTPConflict
	}
	const command = "iam.rbac.action.update"
	payload := struct {
		ID                uuid.UUID
		Name, Description string
		ExpectedRevision  int64
	}{input.ID, input.DisplayName, input.Description, input.ExpectedRevision}
	operationID, requestSHA, _, err := application.mutationContext(
		claims, command, input.ID, input.IdempotencyKey, payload,
	)
	if err != nil {
		return RBACActionView{}, err
	}
	var revision int64
	err = application.inCommand(
		ctx, claims, operationID.String(), "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			result, updateErr := application.runtime.actions.Update(
				scoped, scope, rbacdomain.UpdateCatalog{
					ID: rbacdomain.ID(input.ID), DisplayName: input.DisplayName,
					Description: input.Description, ExpectedRevision: input.ExpectedRevision,
					Evidence: rbacMutationEvidence(claims, permissionID, operationID, requestSHA),
				},
			)
			revision = result.Revision
			return updateErr
		},
	)
	if err != nil {
		return RBACActionView{}, err
	}
	view := rbacActionMutationView(current, input.DisplayName, input.Description, revision)
	if !validRBACActionView(view) {
		return RBACActionView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
