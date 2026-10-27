package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *invitationHTTPApplication) Read(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (InvitationView, error) {
	if application == nil || ctx == nil || ctx.Err() != nil || id == uuid.Nil ||
		!validInvitationClaims(claims) {
		return InvitationView{}, ErrInvitationHTTPInvalid
	}
	request, err := application.snapshotScope(ctx, claims)
	if err != nil {
		return InvitationView{}, err
	}
	var record invitation.Invitation
	err = application.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			var readErr error
			record, readErr = application.runtime.service.Read(
				scoped, scope, invitation.ReadQuery{ID: id},
			)
			return readErr
		},
	)
	if err != nil {
		return InvitationView{}, mapInvitationApplicationError(err)
	}
	return invitationHTTPView(record), nil
}
