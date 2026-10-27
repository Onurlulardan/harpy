package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) UpdateUser(
	ctx context.Context,
	claims accesstoken.Claims,
	input IdentityUserUpdate,
) (IdentityUserView, error) {
	profileChanged := input.FirstName != nil || input.LastName != nil || input.Phone != nil ||
		len(input.CustomAttributes) != 0
	relationsChanged := input.OrganizationID != nil || input.RoleIDs != nil
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 || input.IdempotencyKey == "" ||
		!profileChanged && !relationsChanged && input.LegacyState == "" ||
		input.LegacyState != "" && input.LegacyState != "INACTIVE" {
		return IdentityUserView{}, ErrIdentityHTTPInvalid
	}
	if relationsChanged && input.ID == claims.Subject() {
		return IdentityUserView{}, ErrIdentityHTTPForbidden
	}
	custom, err := identityProfileValues(input.CustomAttributes)
	if err != nil {
		return IdentityUserView{}, err
	}
	operationID := identityOperationID(
		claims, "iam.identity.user.update", input.IdempotencyKey,
	)
	requestSHA, err := identityUserUpdateRequestSHA(input, custom)
	if err != nil {
		return IdentityUserView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.user.update", IdentityUsers, "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			revision := input.ExpectedRevision
			commandKey := "iam.identity.user.update"
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			var replayed bool
			resultSet := false
			if relationsChanged {
				result, replaceErr := application.replaceLegacyUserRelations(
					scoped, scope, claims, permissionID, input, custom,
					profileChanged, evidence,
				)
				if replaceErr != nil {
					return replaceErr
				}
				replayed, resultSet = result.Replayed, true
				revision++
			} else if profileChanged {
				user, getErr := application.identities.users.GetUser(
					scoped, scope, userapp.GetUserQuery{
						ID:       identityfacade.IdentityID(input.ID),
						Evidence: identityQueryEvidence(claims, permissionID),
					},
				)
				if getErr != nil {
					return getErr
				}
				patch, patchErr := application.userProfileUpdateValues(
					scoped, scope, user, custom,
					input.FirstName, input.LastName, input.Phone,
				)
				if patchErr != nil {
					return patchErr
				}
				result, patchErr := application.identities.users.UpdateUser(
					scoped, scope, userapp.UpdateUserCommand{
						ID: identityfacade.IdentityID(input.ID), ExpectedRevision: revision,
						ProfilePatch: patch,
						Evidence:     evidence,
					},
				)
				if patchErr != nil {
					return patchErr
				}
				replayed, resultSet = result.Replayed, true
				revision++
			}
			if input.LegacyState == "" {
				return application.finishIdentityMutation(
					scoped, scope, evidence, input.ID, "user", commandKey, replayed,
				)
			}
			disableOperationID := operationID
			if profileChanged || relationsChanged {
				disableOperationID = identityChildID(operationID, "disable-operation")
			} else {
				commandKey = "iam.identity.user.disable"
				evidence = identityMutationEvidence(
					claims, permissionID, disableOperationID, requestSHA,
				)
			}
			disableEvidence := identityMutationEvidence(
				claims, permissionID, disableOperationID, requestSHA,
			)
			result, disableErr := application.identities.users.DisableUser(
				scoped, scope, userapp.UserStateCommand{
					ID: identityfacade.IdentityID(input.ID), ExpectedRevision: revision,
					Evidence: disableEvidence,
				},
			)
			if disableErr != nil {
				return disableErr
			}
			if resultSet && replayed != result.Replayed {
				return ErrIdentityHTTPUnavailable
			}
			replayed = result.Replayed
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "user", commandKey, replayed,
			)
		},
	)
	if err != nil {
		return IdentityUserView{}, err
	}
	return application.getUser(ctx, claims, input.ID, "update")
}
