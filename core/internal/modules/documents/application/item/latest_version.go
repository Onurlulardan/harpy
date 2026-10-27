package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) LatestVersion(ctx context.Context,
	scope platformtx.TenantScope, itemID domain.ID,
) (Version, error) {
	if !itemID.UUIDv7() {
		return Version{}, ErrInvalid
	}
	store, _, err := application.bind(ctx, scope, SnapshotAccess,
		OperationLatest, PermissionRead)
	if err != nil {
		return Version{}, err
	}
	result, err := store.Get(ctx, itemID)
	if err != nil {
		return Version{}, mapStoreError(err)
	}
	if result.ID != itemID || result.State == domain.ItemPurged ||
		!result.Current.ID.UUIDv7() {
		return Version{}, ErrNotFound
	}
	return result.Current, nil
}
