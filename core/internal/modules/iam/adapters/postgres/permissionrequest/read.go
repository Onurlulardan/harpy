package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) Read(
	ctx context.Context,
	scope tenantScope,
	query permissionrequest.ReadQuery,
) (permissionrequest.Group, error) {
	if query.Validate() != nil ||
		!permissionRequestSnapshotScopeUsable(ctx, scope) {
		return permissionrequest.Group{}, permissionrequest.ErrInvalid
	}
	bound, err := store.bind(scope)
	if err != nil {
		return permissionrequest.Group{}, err
	}
	return scanPermissionRequestGroup(
		bound.database.QueryRow(
			ctx, permissionRequestReadSQL, scope.ActorID(), query.GroupID,
		),
	)
}
