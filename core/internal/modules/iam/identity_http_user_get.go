package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) GetUser(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (IdentityUserView, error) {
	return application.getUser(ctx, claims, id, "read")
}

func (application *identityHTTPApplication) getUser(
	ctx context.Context,
	evidence identityEvidenceSource,
	id uuid.UUID,
	action string,
) (IdentityUserView, error) {
	if id == uuid.Nil {
		return IdentityUserView{}, ErrIdentityHTTPInvalid
	}
	var view IdentityUserView
	err := application.inSnapshotWithAction(
		ctx, evidence, "iam.identity.user.read", IdentityUsers,
		action,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			user, err := application.identities.users.GetUser(
				scoped, scope, userapp.GetUserQuery{
					ID:       identityfacade.IdentityID(id),
					Evidence: identityQueryEvidence(evidence, permissionID),
				},
			)
			if err != nil {
				return err
			}
			view, err = application.userView(scoped, scope, user)
			return err
		},
	)
	return view, err
}
