package iam

import (
	"context"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userrelationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/userrelation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) initializeUserRelations(
	ctx context.Context,
	scope platformtx.TenantScope,
	claims identityEvidenceSource,
	userID uuid.UUID,
	organizations []uuid.UUID,
	grants []identityUserRoleGrant,
	operationID uuid.UUID,
	permissionID uuid.UUID,
	requestSHA [32]byte,
) (bool, error) {
	organizationPatch, err := userrelationapp.NewOrganizationSetPatch(
		identityfacade.IdentityID(userID), true,
		identityFacadeIDs(organizations),
	)
	if err != nil {
		return false, err
	}
	roleIDs := make([]uuid.UUID, len(grants))
	for index := range grants {
		roleIDs[index] = grants[index].roleID
	}
	rolePatch, err := userrelationapp.NewRoleSetPatch(
		identityfacade.IdentityID(userID), len(roleIDs) > 0,
		identityFacadeIDs(roleIDs),
	)
	if err != nil {
		return false, err
	}
	result, err := application.identities.userRelations.Replace(
		ctx, scope, userrelationapp.ReplaceCommand{
			UserID: identityfacade.IdentityID(userID), ExpectedRevision: 1,
			CommandKey:    userrelationapp.CreateCommandKey,
			Organizations: organizationPatch, Roles: rolePatch,
			Evidence: identityMutationEvidence(
				claims, permissionID,
				identityChildID(operationID, "user-relations-initialize-operation"),
				requestSHA,
			),
		},
	)
	if err != nil {
		return false, err
	}
	return result.Replayed, nil
}

func identityFacadeIDs(values []uuid.UUID) []identityfacade.IdentityID {
	result := make([]identityfacade.IdentityID, len(values))
	for index := range values {
		result[index] = identityfacade.IdentityID(values[index])
	}
	return result
}
