package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *portalUserHTTPApplication) GrantPortalConsole(
	ctx context.Context, claims accesstoken.Claims, input PortalUserMutationInput,
) (PortalUserView, error) {
	if !validPortalMutation(input, true) {
		return PortalUserView{}, ErrPortalUserHTTPInvalid
	}
	operationID := identityOperationID(claims, "iam.portal_user.console.grant", input.IdempotencyKey)
	requestSHA, _ := identityRequestSHA(struct {
		ID       uuid.UUID
		Revision int64
	}{
		input.ID, input.ExpectedRevision})
	var view portaluser.View
	err := application.inCommand(ctx, claims, "iam.portal_user.console.grant", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			command := portaluser.RevisionCommand{UserID: input.ID,
				ExpectedRevision: input.ExpectedRevision,
				AccessModeID:     identityChildID(operationID, "console-access-mode"),
				AssignmentID:     identityChildID(operationID, "console-role-assignment"),
				Evidence:         portalMutationEvidence(claims, permissionID, operationID, requestSHA)}
			result, executeErr := application.users.GrantConsole(scoped, scope, command)
			if executeErr != nil {
				return executeErr
			}
			if executeErr = application.finishPortalUserChange(scoped, scope, command.Evidence,
				result.ID, "iam.portal_user.console.grant", result.Replayed); executeErr != nil {
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
