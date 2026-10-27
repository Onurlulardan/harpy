package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) Cancel(
	ctx context.Context,
	scope tenantScope,
	command permissionrequest.CancelCommand,
) (permissionrequest.GroupRef, error) {
	if command.Validate() != nil ||
		!permissionRequestCommandScopeUsable(ctx, scope) ||
		scope.ActorID() != command.RequesterID {
		return permissionrequest.GroupRef{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.GroupRef{}, err
	}
	return scanPermissionRequestReference(bound.database.QueryRow(
		ctx, permissionRequestCancelSQL,
		command.GroupID, command.ExpectedRevision, string(command.ExpectedState),
		command.RequesterID, command.ExpectedAuthorizationVersion,
	))
}
