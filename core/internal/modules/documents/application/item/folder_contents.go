package item

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) FolderContents(ctx context.Context,
	scope platformtx.TenantScope, request ListRequest,
) (Page, error) {
	if !request.FolderID.UUIDv7() || request.Limit < 1 || request.Limit > 100 ||
		(request.After.ID.Valid() && (!request.After.ID.UUIDv7() || request.After.UpdatedAt.IsZero())) {
		return Page{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, SnapshotAccess,
		OperationFolderContents, PermissionFolderRead)
	if err != nil {
		return Page{}, err
	}
	result, err := store.List(ctx, request.FolderID, request.After, request.Limit)
	if err != nil {
		return Page{}, mapStoreError(err)
	}
	if len(result.Items) > int(request.Limit) {
		return Page{}, ErrUnavailable
	}
	return result, nil
}
