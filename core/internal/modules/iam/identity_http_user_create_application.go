package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identityUserCreateIDs struct {
	operation, user, identifier, accessMode uuid.UUID
	defaultAssignment                       uuid.UUID
}

func (application *identityHTTPApplication) CreateUser(
	ctx context.Context,
	claims accesstoken.Claims,
	input IdentityUserCreate,
) (IdentityUserView, error) {
	defer input.Destroy()
	if input.IdempotencyKey == "" || input.OrganizationID == uuid.Nil ||
		input.FirstName == "" || input.LastName == "" {
		return IdentityUserView{}, ErrIdentityHTTPInvalid
	}
	custom, err := identityProfileValues(input.CustomAttributes)
	if err != nil {
		return IdentityUserView{}, err
	}
	grants, err := application.userRoleGrants(
		ctx, claims, input.OrganizationID, input.RoleIDs,
	)
	if err != nil {
		return IdentityUserView{}, err
	}
	credential, err := application.prepareUserCredential(ctx, input)
	if err != nil {
		return IdentityUserView{}, err
	}
	ids := identityNewUserIDs(claims, input.IdempotencyKey)
	requestSHA, err := identityUserCreateRequestSHA(input, custom, credential)
	if err != nil {
		return IdentityUserView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.user.create", IdentityUsers, "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			replayed, createErr := application.createUserGraph(
				scoped, scope, claims, input, custom, grants,
				credential, ids, permissionID, requestSHA,
			)
			if createErr != nil {
				return createErr
			}
			return application.finishIdentityMutation(
				scoped, scope,
				identityMutationEvidence(claims, permissionID, ids.operation, requestSHA),
				ids.user, "user", "iam.identity.user.create", replayed,
			)
		},
	)
	if err != nil {
		return IdentityUserView{}, err
	}
	return application.getUser(ctx, claims, ids.user, "create")
}

func identityNewUserIDs(
	claims identityEvidenceSource, idempotencyKey string,
) identityUserCreateIDs {
	operation := identityOperationID(
		claims, "iam.identity.user.create", idempotencyKey,
	)
	return identityUserCreateIDs{
		operation: operation, user: identityChildID(operation, "user"),
		identifier:        identityChildID(operation, "email-identifier"),
		accessMode:        identityChildID(operation, "operator-access-mode"),
		defaultAssignment: identityChildID(operation, "default-role-assignment"),
	}
}

func identityUserCreateRequestSHA(
	input IdentityUserCreate,
	custom []byte,
	credential identityPreparedUserCredential,
) ([32]byte, error) {
	return identityRequestSHA(struct {
		Email, FirstName, LastName, Phone string
		OrganizationID                    uuid.UUID
		RoleIDs                           []uuid.UUID
		Custom                            []byte
		PasswordSHA256                    [32]byte
	}{
		input.Email, input.FirstName, input.LastName, input.Phone, input.OrganizationID,
		identitySortedIDs(input.RoleIDs), custom, credential.passwordSHA256,
	})
}
