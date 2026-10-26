package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *portalUserHTTPApplication) InvitePortalUser(
	ctx context.Context, claims accesstoken.Claims, input PortalUserMutationInput,
) (PortalUserView, error) {
	if !validPortalMutation(input, false) {
		return PortalUserView{}, ErrPortalUserHTTPInvalid
	}
	var view portaluser.View
	err := application.inSnapshot(ctx, claims, "update",
		func(scoped context.Context, scope platformtx.TenantScope, _ uuid.UUID) error {
			var readErr error
			view, readErr = application.users.Read(scoped, scope, input.ID)
			return readErr
		})
	if err != nil {
		return PortalUserView{}, err
	}
	err = application.setups.StartPortalSetup(ctx, claims, LifecycleResourceInput{
		ID: view.BindingID, IdempotencyKey: input.IdempotencyKey,
	})
	if err != nil {
		return PortalUserView{}, mapPortalLifecycleError(err)
	}
	result := portalUserView(view)
	result.InviteSent = true
	return result, nil
}

func mapPortalLifecycleError(err error) error {
	switch err {
	case nil:
		return nil
	case ErrLifecycleHTTPInvalid:
		return ErrPortalUserHTTPInvalid
	case ErrLifecycleHTTPForbidden:
		return ErrPortalUserHTTPForbidden
	case ErrLifecycleHTTPNotFound:
		return ErrPortalUserHTTPNotFound
	case ErrLifecycleHTTPConflict:
		return ErrPortalUserHTTPConflict
	default:
		return ErrPortalUserHTTPUnavailable
	}
}
