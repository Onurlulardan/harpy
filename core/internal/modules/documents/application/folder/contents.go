package folder

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Contents(ctx context.Context,
	scope platformtx.TenantScope, request ListRequest,
) (Page, error) {
	if !request.ParentID.UUIDv7() || request.AfterID.Valid() && !request.AfterID.UUIDv7() ||
		request.Limit < 1 || request.Limit > 100 {
		return Page{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, Snapshot,
		OperationContents, PermissionRead)
	if err != nil {
		return Page{}, err
	}
	return listBound(ctx, store, request)
}
