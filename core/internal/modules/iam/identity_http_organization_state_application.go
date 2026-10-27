package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) ChangeOrganizationState(
	ctx context.Context, claims accesstoken.Claims, input IdentityStateMutation,
) (IdentityOrganizationView, error) {
	if input.State != "SUSPENDED" {
		return IdentityOrganizationView{}, ErrIdentityHTTPInvalid
	}
	operationID, requestSHA, err := identityStateContract(
		claims, "iam.identity.organization.disable", input,
	)
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.organization.disable", IdentityOrganizations, "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, disableErr := application.identities.organizations.DisableOrganization(
				scoped, scope, organizationapp.OrganizationStateCommand{
					ID:               identityfacade.IdentityID(input.ID),
					ExpectedRevision: input.ExpectedRevision,
					Evidence:         evidence,
				},
			)
			if disableErr != nil {
				return disableErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "organization",
				"iam.identity.organization.disable", result.Replayed,
			)
		},
	)
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	return application.getOrganization(ctx, claims, input.ID, "update")
}
