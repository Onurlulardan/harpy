package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *permissionRequestHTTPApplication) Approve(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestDecision,
) (PermissionRequestRefView, error) {
	return application.decide(ctx, claims, input, permissionrequest.DecisionApprove)
}

func (application *permissionRequestHTTPApplication) Reject(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestDecision,
) (PermissionRequestRefView, error) {
	return application.decide(ctx, claims, input, permissionrequest.DecisionReject)
}

func (application *permissionRequestHTTPApplication) decide(
	ctx context.Context,
	claims accesstoken.Claims,
	input PermissionRequestDecision,
	decision permissionrequest.Decision,
) (PermissionRequestRefView, error) {
	if input.ID == uuid.Nil || input.ReviewTaskID == uuid.Nil ||
		input.ExpectedRevision < 1 || input.ExpectedState != permissionrequest.StateInReview ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return PermissionRequestRefView{}, ErrPermissionRequestHTTPInvalid
	}
	operationID := permissionRequestOperationID(
		claims, "iam.permission_requests."+stringLowerDecision(decision), input.IdempotencyKey,
	)
	requestSHA, err := permissionRequestRequestSHA(struct {
		Input    PermissionRequestDecision
		Decision permissionrequest.Decision
	}{input, decision})
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	stepUp, valid := permissionRequestStepUp(claims, input.ID)
	if !valid {
		return PermissionRequestRefView{}, ErrPermissionRequestHTTPForbidden
	}
	command := permissionrequest.DecideCommand{
		OperationID: operationID, RequestSHA256: requestSHA,
		GroupID: input.ID, ExpectedRevision: input.ExpectedRevision,
		ExpectedState: input.ExpectedState, ReviewerID: claims.Subject(),
		ExpectedReviewerTokenVersion:         claims.TokenVersion(),
		ExpectedReviewerAuthorizationVersion: claims.AuthorizationVersion(),
		ReviewTaskID:                         input.ReviewTaskID, Decision: decision, StepUp: stepUp,
	}
	var result permissionrequest.DecisionResult
	err = application.inCommand(
		ctx, claims, operationID.String(), "review",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var executeErr error
			result, executeErr = application.runtime.decide(scoped, scope, command)
			return executeErr
		},
	)
	if err != nil {
		return PermissionRequestRefView{}, err
	}
	state := permissionrequest.StateApproved
	if decision == permissionrequest.DecisionReject {
		state = permissionrequest.StateRejected
	}
	return permissionRequestRefView(claims.Tenant(), permissionrequest.GroupRef{
		ID: input.ID, State: state, Revision: result.Revision,
	}), nil
}

func permissionRequestStepUp(
	claims accesstoken.Claims,
	groupID uuid.UUID,
) (permissionrequest.StepUpEvidence, bool) {
	purpose, present := claims.StepUpPurposeHash()
	want := permissionRequestPurpose(claims.Tenant(), groupID)
	if !present || purpose != want || claims.Assurance() != accesstoken.AssuranceStepUp ||
		claims.AuthenticationTimeUnix() < 1 ||
		claims.AuthenticationTimeUnix() > claims.IssuedAtUnix() ||
		claims.IssuedAtUnix()-claims.AuthenticationTimeUnix() > 600 {
		return permissionrequest.StepUpEvidence{}, false
	}
	evidence := permissionrequest.StepUpEvidence{
		PurposeSHA256: purpose, AuthTime: time.Unix(claims.AuthenticationTimeUnix(), 0).UTC(),
		ACR: permissionrequest.StepUpACR,
	}
	for _, method := range claims.Methods() {
		evidence.AMRPassword = evidence.AMRPassword || method == accesstoken.AuthMethodPassword
		evidence.AMRWebAuthn = evidence.AMRWebAuthn || method == accesstoken.AuthMethodWebAuthn
		evidence.AMRTOTP = evidence.AMRTOTP || method == accesstoken.AuthMethodTOTP
	}
	return evidence, evidence.Validate() == nil
}

func stringLowerDecision(decision permissionrequest.Decision) string {
	if decision == permissionrequest.DecisionApprove {
		return "approve"
	}
	return "reject"
}
