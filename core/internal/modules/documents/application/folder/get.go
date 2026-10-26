package folder

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Get(ctx context.Context, scope platformtx.TenantScope,
	request GetRequest,
) (domain.Folder, error) {
	if !request.FolderID.UUIDv7() {
		return domain.Folder{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, Snapshot, OperationGet, PermissionRead)
	if err != nil {
		return domain.Folder{}, err
	}
	return getBound(ctx, store, request.FolderID)
}

func getBound(ctx context.Context, store Store, id domain.ID) (domain.Folder, error) {
	result, err := store.Get(ctx, id)
	if err != nil {
		return domain.Folder{}, mapStoreError(err)
	}
	return result, nil
}
