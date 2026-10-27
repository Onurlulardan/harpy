package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) Review(
	ctx context.Context,
	scope tenantScope,
	command permissionrequest.ReviewCommand,
) (permissionrequest.GroupRef, error) {
	if command.Validate() != nil ||
		!permissionRequestCommandScopeUsable(ctx, scope) ||
		scope.ActorID() != command.ReviewerID {
		return permissionrequest.GroupRef{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.GroupRef{}, err
	}
	return scanPermissionRequestReference(bound.database.QueryRow(
		ctx, permissionRequestReviewSQL,
		command.GroupID, command.ExpectedRevision, string(command.ExpectedState),
		command.ReviewerID, command.ExpectedReviewerAuthorizationVersion,
		command.ReviewTaskID,
	))
}
