package iam

import (
	"context"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	passwordenrollment "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/passwordenrollment"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) createUserGraph(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims identityEvidenceSource,
	input IdentityUserCreate,
	custom []byte,
	grants []identityUserRoleGrant,
	credential identityPreparedUserCredential,
	ids identityUserCreateIDs,
	permissionID uuid.UUID,
	requestSHA [32]byte,
) (bool, error) {
	profile, err := application.userProfileValues(
		ctx, scope, custom, &input.FirstName, &input.LastName, &input.Phone,
	)
	if err != nil {
		return false, err
	}
	return application.createUserGraphPrepared(
		ctx, scope, claims, []uuid.UUID{input.OrganizationID}, profile, grants,
		credential, ids, permissionID, requestSHA,
	)
}

func (application *identityHTTPApplication) createUserGraphPrepared(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims identityEvidenceSource,
	organizationIDs []uuid.UUID,
	profile []byte,
	grants []identityUserRoleGrant,
	credential identityPreparedUserCredential,
	ids identityUserCreateIDs,
	permissionID uuid.UUID,
	requestSHA [32]byte,
) (bool, error) {
	result, err := application.identities.users.CreateUser(
		ctx, scope, userapp.CreateUserCommand{
			ID:                  identityfacade.IdentityID(ids.user),
			IdentifierID:        identityfacade.IdentityID(ids.identifier),
			AccessModeID:        identityfacade.IdentityID(ids.accessMode),
			DefaultAssignmentID: identityfacade.IdentityID(ids.defaultAssignment),
			Status:              identityfacade.UserActive, ProfileValues: profile,
			Directory: credential.directory,
			Evidence: identityMutationEvidence(
				claims, permissionID, ids.operation, requestSHA,
			),
		},
	)
	if err != nil {
		return false, err
	}
	if err = application.createUserPassword(
		ctx, scope, claims, credential, ids, permissionID, requestSHA,
	); err != nil {
		return false, err
	}
	relationsReplayed, err := application.initializeUserRelations(
		ctx, scope, claims, ids.user, organizationIDs, grants,
		ids.operation, permissionID, requestSHA,
	)
	if err != nil {
		return false, err
	}
	if relationsReplayed != result.Replayed {
		return false, ErrIdentityHTTPInvalid
	}
	return result.Replayed, nil
}

func (application *identityHTTPApplication) createUserPassword(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims identityEvidenceSource,
	credential identityPreparedUserCredential,
	ids identityUserCreateIDs,
	permissionID uuid.UUID,
	requestSHA [32]byte,
) error {
	_, err := application.identities.passwords.Create(
		ctx, scope, passwordenrollment.Command{
			UserID:     identityfacade.IdentityID(ids.user),
			Credential: credential.password,
			Evidence: identityMutationEvidence(
				claims, permissionID,
				identityChildID(ids.operation, "password-enrollment-operation"), requestSHA,
			),
		},
	)
	return err
}
