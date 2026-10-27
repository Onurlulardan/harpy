package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) CreateResource(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACCatalogCreate,
) (RBACResourceView, error) {
	if !rbacCatalogCreateValid(input) {
		return RBACResourceView{}, ErrRBACHTTPInvalid
	}
	const command = "iam.rbac.resource.create"
	payload := struct{ Key, Name, Description string }{
		input.Key, input.DisplayName, input.Description,
	}
	operationID, requestSHA, _, err := application.mutationContext(
		claims, command, uuid.Nil, input.IdempotencyKey, payload,
	)
	if err != nil {
		return RBACResourceView{}, err
	}
	id := identityChildID(operationID, "rbac-resource")
	var revision int64
	err = application.inCommand(
		ctx, claims, operationID.String(), "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			result, createErr := application.runtime.resources.Create(
				scoped, scope, rbacdomain.ResourceInput{
					ID: rbacdomain.ID(id), Key: input.Key,
					OwnerDomain: "legacy", Kind: "legacy_resource",
					DisplayName: input.DisplayName, Description: input.Description,
					Evidence: rbacMutationEvidence(claims, permissionID, operationID, requestSHA),
				},
			)
			revision = result.Revision
			return createErr
		},
	)
	if err != nil {
		return RBACResourceView{}, err
	}
	view := RBACResourceView{
		ID: id.String(), Name: input.DisplayName, Slug: input.Key,
		Description: input.Description, OwnerDomain: "legacy",
		Kind: "legacy_resource", State: "active", Revision: revision,
	}
	if !validRBACResourceView(view) {
		return RBACResourceView{}, ErrRBACHTTPUnavailable
	}
	return view, nil
}
