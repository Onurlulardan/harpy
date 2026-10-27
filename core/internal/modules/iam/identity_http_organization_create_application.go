package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) CreateOrganization(
	ctx context.Context,
	claims accesstoken.Claims,
	input IdentityOrganizationCreate,
) (IdentityOrganizationView, error) {
	profile, err := identityProfileValues(input.CustomAttributes)
	if err != nil || input.IdempotencyKey == "" || input.Name == "" || input.Slug == "" ||
		len(input.OwnerIDs) == 0 {
		return IdentityOrganizationView{}, ErrIdentityHTTPInvalid
	}
	operationID := identityOperationID(
		claims, "iam.identity.organization.create", input.IdempotencyKey,
	)
	organizationID := identityChildID(operationID, "organization")
	requestSHA, err := identityRequestSHA(struct {
		Name, Slug string
		ParentID   uuid.UUID
		OwnerIDs   []uuid.UUID
		Profile    []byte
	}{input.Name, input.Slug, input.ParentID, identitySortedIDs(input.OwnerIDs), profile})
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.organization.create", IdentityOrganizations, "create",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			ownerIDs, membershipIDs := identityOrganizationOwnerCommand(
				operationID, input.OwnerIDs,
			)
			storedProfile, profileErr := application.profileDocumentValues(
				scoped, scope, identityfacade.ProfileOrganization, profile,
			)
			if profileErr != nil {
				return profileErr
			}
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, createErr := application.identities.organizations.CreateOrganization(
				scoped, scope, organizationapp.CreateOrganizationCommand{
					ID:   identityfacade.IdentityID(organizationID),
					Name: input.Name, Slug: input.Slug,
					ParentID:           identityfacade.IdentityID(input.ParentID),
					OwnerIDs:           ownerIDs,
					OwnerMembershipIDs: membershipIDs,
					ProfileValues:      storedProfile,
					Evidence:           evidence,
				},
			)
			if createErr != nil {
				return createErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, organizationID, "organization",
				"iam.identity.organization.create", result.Replayed,
			)
		},
	)
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	return application.getOrganization(ctx, claims, organizationID, "create")
}
