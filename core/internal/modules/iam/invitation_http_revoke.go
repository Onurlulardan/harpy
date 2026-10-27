package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *invitationHTTPApplication) Revoke(
	ctx context.Context,
	claims accesstoken.Claims,
	input InvitationRevokeInput,
) error {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!validInvitationClaims(claims) || input.ID == [16]byte{} ||
		input.ExpectedRevision < 1 ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrInvitationHTTPInvalid
	}
	operationID := identityOperationID(
		claims, "iam.invitation.revoke", input.IdempotencyKey,
	)
	requestSHA, err := identityRequestSHA(struct {
		ID       [16]byte
		Revision int64
	}{input.ID, input.ExpectedRevision})
	if err != nil {
		return ErrInvitationHTTPInvalid
	}
	request, err := application.commandScope(ctx, claims, "iam.invitation.revoke")
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			_, revokeErr := application.runtime.service.Revoke(
				scoped, scope, invitation.RevokeCommand{
					OperationID: operationID, RequestSHA256: requestSHA,
					ID: input.ID, ExpectedRevision: input.ExpectedRevision,
					ActorID:                      claims.Subject(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
				},
			)
			return revokeErr
		},
	)
	return mapInvitationApplicationError(err)
}
