package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) DeleteUser(
	ctx context.Context, claims accesstoken.Claims, input IdentityMutation,
) error {
	operationID, requestSHA, err := identityMutationContract(
		claims, "iam.identity.user.delete", input,
	)
	if err != nil {
		return err
	}
	return application.inCommand(
		ctx, claims, "iam.identity.user.delete", IdentityUsers, "delete",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, deleteErr := application.identities.users.DeleteUser(
				scoped, scope, userapp.UserStateCommand{
					ID:               identityfacade.IdentityID(input.ID),
					ExpectedRevision: input.ExpectedRevision,
					Evidence:         evidence,
				},
			)
			if deleteErr != nil {
				return deleteErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "user",
				"iam.identity.user.delete", result.Replayed,
			)
		},
	)
}
