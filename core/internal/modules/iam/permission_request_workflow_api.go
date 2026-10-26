package iam

import (
	"context"

	permissionrequestapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/permissionrequest"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *PermissionRequestRuntime) DecidePermissionRequest(
	ctx context.Context,
	scope platformtx.TenantScope,
	command permissionrequestapi.Command,
) (permissionrequestapi.Result, error) {
	operationID, operationValid := permissionRequestAPIUUID(command.OperationID)
	groupID, groupValid := permissionRequestAPIUUID(command.GroupID)
	reviewerID, reviewerValid := permissionRequestAPIUUID(command.ReviewerID)
	taskID, taskValid := permissionRequestAPIUUID(command.ReviewTaskID)
	if !operationValid || !groupValid || !reviewerValid || !taskValid {
		return permissionrequestapi.Result{}, permissionrequest.ErrInvalid
	}
	result, err := runtime.decide(ctx, scope, permissionrequest.DecideCommand{
		OperationID: operationID, RequestSHA256: command.RequestSHA256,
		GroupID: groupID, ExpectedRevision: command.ExpectedRevision,
		ExpectedState:                        permissionrequest.StateInReview,
		ReviewerID:                           reviewerID,
		ExpectedReviewerTokenVersion:         command.ExpectedReviewerTokenVersion,
		ExpectedReviewerAuthorizationVersion: command.ExpectedReviewerAuthorizationVersion,
		ReviewTaskID:                         taskID,
		Decision:                             permissionrequest.Decision(command.Decision),
		StepUp: permissionrequest.StepUpEvidence{
			PurposeSHA256: command.StepUp.PurposeSHA256,
			AuthTime:      command.StepUp.AuthTime, ACR: command.StepUp.ACR,
			AMRPassword: command.StepUp.AMRPassword,
			AMRWebAuthn: command.StepUp.AMRWebAuthn,
			AMRTOTP:     command.StepUp.AMRTOTP,
		},
	})
	if err != nil {
		return permissionrequestapi.Result{}, err
	}
	return permissionrequestapi.Result{
		Code: result.Code, Revision: result.Revision,
		AuthorizationClock: result.AuthorizationClock, Replayed: result.Replayed,
	}, nil
}

func permissionRequestAPIUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value
}

var _ permissionrequestapi.Decider = (*PermissionRequestRuntime)(nil)
