package upload

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Get(ctx context.Context, scope platformtx.TenantScope,
	request GetRequest,
) (Snapshot, error) {
	if !request.UploadID.UUIDv7() {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, SnapshotAccess, OperationGet, PermissionRead)
	if err != nil {
		return Snapshot{}, err
	}
	return getBound(ctx, store, request.UploadID)
}

func getBound(ctx context.Context, store Store, id domain.ID) (Snapshot, error) {
	result, err := store.Get(ctx, id)
	if err != nil {
		return Snapshot{}, mapStoreError(err)
	}
	return result, nil
}
