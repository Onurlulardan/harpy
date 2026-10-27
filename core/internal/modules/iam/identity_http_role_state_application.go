package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) ChangeRoleState(
	ctx context.Context, claims accesstoken.Claims, input IdentityStateMutation,
) (IdentityRoleView, error) {
	if input.State != "RETIRED" {
		return IdentityRoleView{}, ErrIdentityHTTPInvalid
	}
	operationID, requestSHA, err := identityStateContract(
		claims, "iam.identity.role.retire", input,
	)
	if err != nil {
		return IdentityRoleView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.role.retire", IdentityRoles, "delete",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, retireErr := application.identities.roles.RetireRole(
				scoped, scope, roleapp.RetireRoleCommand{
					ID:               identityfacade.IdentityID(input.ID),
					ExpectedRevision: input.ExpectedRevision,
					Evidence:         evidence,
				},
			)
			if retireErr != nil {
				return retireErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "role",
				"iam.identity.role.retire", result.Replayed,
			)
		},
	)
	if err != nil {
		return IdentityRoleView{}, err
	}
	return application.getRole(ctx, claims, input.ID, "delete")
}
