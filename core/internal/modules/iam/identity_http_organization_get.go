package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	organizationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/organization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) GetOrganization(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (IdentityOrganizationView, error) {
	return application.getOrganization(ctx, claims, id, "read")
}

func (application *identityHTTPApplication) getOrganization(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
	action string,
) (IdentityOrganizationView, error) {
	if id == uuid.Nil {
		return IdentityOrganizationView{}, ErrIdentityHTTPInvalid
	}
	var view IdentityOrganizationView
	err := application.inSnapshotWithAction(
		ctx, claims, "iam.identity.organization.read", IdentityOrganizations,
		action,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			organization, err := application.identities.organizations.GetOrganization(
				scoped, scope, organizationapp.GetOrganizationQuery{
					ID:       identityfacade.IdentityID(id),
					Evidence: identityQueryEvidence(claims, permissionID),
				},
			)
			if err == nil {
				view, err = application.organizationView(scoped, scope, organization)
			}
			return err
		},
	)
	return view, err
}
