package item

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) ReadVersions(ctx context.Context,
	scope platformtx.TenantScope, request ListVersionsRequest,
) (VersionPage, error) {
	if !request.ItemID.UUIDv7() || request.Limit == 0 || request.Limit > 100 {
		return VersionPage{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, SnapshotAccess,
		OperationListVersions, PermissionRead)
	if err != nil {
		return VersionPage{}, err
	}
	result, err := store.ReadVersions(ctx, request.ItemID,
		request.AfterVersionNumber, request.Limit)
	if err != nil {
		return VersionPage{}, mapStoreError(err)
	}
	if len(result.Items) > int(request.Limit) {
		return VersionPage{}, ErrUnavailable
	}
	return result, nil
}
