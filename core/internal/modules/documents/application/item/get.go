package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Get(ctx context.Context, scope platformtx.TenantScope,
	request GetRequest,
) (Snapshot, error) {
	if !request.ItemID.UUIDv7() {
		return Snapshot{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, SnapshotAccess, OperationGet, PermissionRead)
	if err != nil {
		return Snapshot{}, err
	}
	result, err := store.Get(ctx, request.ItemID)
	if err != nil {
		return Snapshot{}, mapStoreError(err)
	}
	if result.ID != request.ItemID || result.State == domain.ItemPurged {
		return Snapshot{}, ErrNotFound
	}
	return result, nil
}
