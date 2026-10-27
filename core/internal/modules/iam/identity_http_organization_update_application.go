package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) UpdateOrganization(
	ctx context.Context,
	claims accesstoken.Claims,
	input IdentityOrganizationUpdate,
) (IdentityOrganizationView, error) {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 || input.IdempotencyKey == "" {
		return IdentityOrganizationView{}, ErrIdentityHTTPInvalid
	}
	current, err := application.getOrganization(ctx, claims, input.ID, "update")
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	name, slug, parentID := current.Name, current.Slug, current.ParentID
	if input.Name != nil {
		name = *input.Name
	}
	if input.Slug != nil {
		slug = *input.Slug
	}
	if input.ParentID != nil {
		parentID = *input.ParentID
	}
	patch, err := identityProfileValues(input.CustomAttributes)
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	operationID := identityOperationID(
		claims, "iam.identity.organization.update", input.IdempotencyKey,
	)
	var ownerIDs []uuid.UUID
	if input.OwnerIDs != nil {
		ownerIDs = identitySortedIDs(*input.OwnerIDs)
	}
	requestSHA, err := identityRequestSHA(struct {
		ID               uuid.UUID
		ExpectedRevision int64
		Name, Slug       string
		ParentID         uuid.UUID
		ReplaceOwners    bool
		OwnerIDs         []uuid.UUID
		Patch            []byte
	}{input.ID, input.ExpectedRevision, name, slug, parentID,
		input.OwnerIDs != nil, ownerIDs, patch})
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	err = application.inCommand(
		ctx, claims, "iam.identity.organization.update", IdentityOrganizations, "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			commandOwners, membershipIDs := identityOrganizationOwnerCommand(
				operationID, ownerIDs,
			)
			storedPatch, profileErr := application.profileDocumentValues(
				scoped, scope, identityfacade.ProfileOrganization, patch,
			)
			if profileErr != nil {
				return profileErr
			}
			evidence := identityMutationEvidence(
				claims, permissionID, operationID, requestSHA,
			)
			result, updateErr := application.identities.organizations.UpdateOrganization(
				scoped, scope, organizationapp.UpdateOrganizationCommand{
					ID:               identityfacade.IdentityID(input.ID),
					ExpectedRevision: input.ExpectedRevision,
					Name:             name, Slug: slug, ParentID: identityfacade.IdentityID(parentID),
					ReplaceOwners:      input.OwnerIDs != nil,
					OwnerIDs:           commandOwners,
					OwnerMembershipIDs: membershipIDs,
					ProfilePatch:       storedPatch,
					Evidence:           evidence,
				},
			)
			if updateErr != nil {
				return updateErr
			}
			return application.finishIdentityMutation(
				scoped, scope, evidence, input.ID, "organization",
				"iam.identity.organization.update", result.Replayed,
			)
		},
	)
	if err != nil {
		return IdentityOrganizationView{}, err
	}
	return application.getOrganization(ctx, claims, input.ID, "update")
}
