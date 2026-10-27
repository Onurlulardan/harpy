package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) Decide(
	ctx context.Context,
	scope tenantScope,
	command permissionrequest.DecideCommand,
) (permissionrequest.DecisionResult, error) {
	if command.Validate() != nil ||
		!permissionRequestCommandScopeUsable(ctx, scope) ||
		scope.ActorID() != command.ReviewerID {
		return permissionrequest.DecisionResult{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.DecisionResult{}, err
	}
	var result permissionrequest.DecisionResult
	err = bound.database.QueryRow(
		ctx, permissionRequestDecideSQL,
		command.OperationID, command.RequestSHA256[:], command.GroupID,
		command.ExpectedRevision, string(command.ExpectedState), command.ReviewerID,
		command.ExpectedReviewerTokenVersion,
		command.ExpectedReviewerAuthorizationVersion,
		command.ReviewTaskID, string(command.Decision),
		command.StepUp.PurposeSHA256[:], command.StepUp.AuthTime,
		command.StepUp.ACR, command.StepUp.AMRPassword,
		command.StepUp.AMRWebAuthn, command.StepUp.AMRTOTP,
	).Scan(
		&result.Code, &result.Revision, &result.AuthorizationClock, &result.Replayed,
	)
	if err != nil {
		return permissionrequest.DecisionResult{}, mapPermissionRequestError(err)
	}
	if result.Revision < 1 || result.AuthorizationClock < 1 || result.Code == "" {
		return permissionrequest.DecisionResult{}, permissionrequest.ErrUnavailable
	}
	return result, nil
}
