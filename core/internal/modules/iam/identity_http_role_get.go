package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) GetRole(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (IdentityRoleView, error) {
	return application.getRole(ctx, claims, id, "read")
}

func (application *identityHTTPApplication) getRole(
	ctx context.Context,
	evidence identityEvidenceSource,
	id uuid.UUID,
	action string,
) (IdentityRoleView, error) {
	if id == uuid.Nil {
		return IdentityRoleView{}, ErrIdentityHTTPInvalid
	}
	var view IdentityRoleView
	err := application.inSnapshotWithAction(
		ctx, evidence, "iam.identity.role.read", IdentityRoles,
		action,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			role, err := application.identities.roles.GetRole(
				scoped, scope, roleapp.GetRoleQuery{
					ID:       identityfacade.IdentityID(id),
					Evidence: identityQueryEvidence(evidence, permissionID),
				},
			)
			if err == nil {
				view, err = application.roleView(scoped, scope, role)
			}
			return err
		},
	)
	return view, err
}
