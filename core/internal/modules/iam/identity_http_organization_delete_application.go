package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) DeleteOrganization(
	ctx context.Context, claims accesstoken.Claims, input IdentityMutation,
) error {
	operationID, requestSHA, err := identityMutationContract(
		claims, "iam.identity.organization.delete", input,
	)
	if err != nil {
		return err
	}
	return application.inCommand(
		ctx, claims, "iam.identity.organization.delete", IdentityOrganizations, "delete",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, deleteErr := application.identities.organizations.DeleteOrganization(
				scoped, scope, organizationapp.OrganizationStateCommand{
					ID:               identityfacade.IdentityID(input.ID),
					ExpectedRevision: input.ExpectedRevision,
					Evidence:         evidence,
				},
			)
			if deleteErr != nil {
				return deleteErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "organization",
				"iam.identity.organization.delete", result.Replayed,
			)
		},
	)
}
