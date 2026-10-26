package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *permissionRequestHTTPApplication) Review(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestReview,
) (PermissionRequestRefView, error) {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 ||
		input.ExpectedState != permissionrequest.StatePending ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return PermissionRequestRefView{}, ErrPermissionRequestHTTPInvalid
	}
	operationID := permissionRequestOperationID(
		claims, "iam.permission_requests.review", input.IdempotencyKey,
	)
	requestSHA, err := permissionRequestRequestSHA(input)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	command := permissionrequest.ReviewCommand{
		OperationID: operationID, RequestSHA256: requestSHA,
		GroupID: input.ID, ExpectedRevision: input.ExpectedRevision,
		ExpectedState: input.ExpectedState, ReviewerID: claims.Subject(),
		ExpectedReviewerAuthorizationVersion: claims.AuthorizationVersion(),
		ReviewTaskID:                         permissionRequestChildID(operationID, "review-task"),
	}
	var result permissionrequest.GroupRef
	err = application.inCommand(
		ctx, claims, operationID.String(), "review",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var executeErr error
			result, executeErr = application.runtime.service.Review(scoped, scope, command)
			return executeErr
		},
	)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	return permissionRequestRefView(claims.Tenant(), result), nil
}
