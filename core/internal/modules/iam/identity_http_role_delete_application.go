package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) DeleteRole(
	ctx context.Context, claims accesstoken.Claims, input IdentityMutation,
) error {
	operationID, requestSHA, err := identityMutationContract(
		claims, "iam.identity.role.retire", input,
	)
	if err != nil {
		return err
	}
	return application.inCommand(
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
}
