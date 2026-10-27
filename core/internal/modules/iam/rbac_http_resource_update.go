package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) UpdateResource(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACCatalogUpdate,
) (RBACResourceView, error) {
	if !rbacCatalogUpdateValid(input) {
		return RBACResourceView{}, ErrRBACHTTPInvalid
	}
	current, err := application.readResource(ctx, claims, input.ID)
	if err != nil {
		return RBACResourceView{}, err
	}
	if input.Key != "" && input.Key != current.Key {
		return RBACResourceView{}, ErrRBACHTTPConflict
	}
	const command = "iam.rbac.resource.update"
	payload := struct {
		ID                uuid.UUID
		Name, Description string
		ExpectedRevision  int64
	}{input.ID, input.DisplayName, input.Description, input.ExpectedRevision}
	operationID, requestSHA, _, err := application.mutationContext(
		claims, command, input.ID, input.IdempotencyKey, payload,
	)
	if err != nil {
		return RBACResourceView{}, err
	}
	var revision int64
	err = application.inCommand(
		ctx, claims, operationID.String(), "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			result, updateErr := application.runtime.resources.Update(
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
		return RBACResourceView{}, err
	}
	view := rbacCatalogMutationView(current, input.DisplayName, input.Description, revision)
	if !validRBACResourceView(view) {
		return RBACResourceView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
