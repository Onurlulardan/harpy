package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) ChangeUserState(
	ctx context.Context, claims accesstoken.Claims, input IdentityStateMutation,
) (IdentityUserView, error) {
	if input.State != "INACTIVE" {
		return IdentityUserView{}, ErrIdentityHTTPInvalid
	}
	operationID, requestSHA, err := identityStateContract(
		claims, "iam.identity.user.disable", input,
	)
	if err != nil {
		return IdentityUserView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.user.disable", IdentityUsers, "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, disableErr := application.identities.users.DisableUser(
				scoped, scope, userapp.UserStateCommand{
					ID:               identityfacade.IdentityID(input.ID),
					ExpectedRevision: input.ExpectedRevision,
					Evidence:         evidence,
				},
			)
			if disableErr != nil {
				return disableErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "user",
				"iam.identity.user.disable", result.Replayed,
			)
		},
	)
	if err != nil {
		return IdentityUserView{}, err
	}
	return application.getUser(ctx, claims, input.ID, "update")
}
