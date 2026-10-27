package folder

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) List(ctx context.Context, scope platformtx.TenantScope,
	request ListRequest,
) (Page, error) {
	if request.ParentID.Valid() && !request.ParentID.UUIDv7() ||
		request.AfterID.Valid() && !request.AfterID.UUIDv7() || request.Limit < 1 || request.Limit > 200 {
		return Page{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, Snapshot, OperationList, PermissionRead)
	if err != nil {
		return Page{}, err
	}
	return listBound(ctx, store, request)
}

func listBound(ctx context.Context, store Store, request ListRequest) (Page, error) {
	result, err := store.List(ctx, request.ParentID, request.AfterID, request.Limit)
	if err != nil {
		return Page{}, mapStoreError(err)
	}
	return result, nil
}
