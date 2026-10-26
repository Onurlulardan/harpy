package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) Expire(
	ctx context.Context,
	scope tenantScope,
	command permissionrequest.ExpireCommand,
) (permissionrequest.GroupRef, error) {
	if command.Validate() != nil ||
		!permissionRequestCommandScopeUsable(ctx, scope) {
		return permissionrequest.GroupRef{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.GroupRef{}, err
	}
	return scanPermissionRequestReference(bound.database.QueryRow(
		ctx, permissionRequestExpireSQL,
		command.GroupID, command.ExpectedRevision, string(command.ExpectedState),
	))
}
