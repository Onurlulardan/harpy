package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *portalUserHTTPApplication) RevokePortalConsole(
	ctx context.Context, claims accesstoken.Claims, input PortalUserMutationInput,
) (PortalUserView, error) {
	if !validPortalMutation(input, true) {
		return PortalUserView{}, ErrPortalUserHTTPInvalid
	}
	operationID := identityOperationID(claims, "iam.portal_user.console.revoke", input.IdempotencyKey)
	requestSHA, _ := identityRequestSHA(struct {
		ID       uuid.UUID
		Revision int64
	}{
		input.ID, input.ExpectedRevision})
	var view portaluser.View
	err := application.inCommand(ctx, claims, "iam.portal_user.console.revoke", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			command := portaluser.RevisionCommand{UserID: input.ID,
				ExpectedRevision: input.ExpectedRevision,
				Evidence:         portalMutationEvidence(claims, permissionID, operationID, requestSHA)}
			result, executeErr := application.users.RevokeConsole(scoped, scope, command)
			if executeErr != nil {
				return executeErr
			}
			if executeErr = application.finishPortalUserChange(scoped, scope, command.Evidence,
				result.ID, "iam.portal_user.console.revoke", result.Replayed); executeErr != nil {
				return executeErr
			}
			view, executeErr = application.users.Read(scoped, scope, result.ID)
			return executeErr
		})
	if err != nil {
		return PortalUserView{}, err
	}
	return portalUserView(view), nil
}
