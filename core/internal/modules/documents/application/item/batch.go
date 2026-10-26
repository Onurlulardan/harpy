package item

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Batch(ctx context.Context,
	scope platformtx.TenantScope, itemIDs []domain.ID,
) ([]Snapshot, error) {
	if len(itemIDs) < 1 || len(itemIDs) > 100 {
		return nil, ErrInvalid
	}
	seen := make(map[domain.ID]struct{}, len(itemIDs))
	for _, itemID := range itemIDs {
		if !itemID.UUIDv7() {
			return nil, ErrInvalid
		}
		if _, duplicate := seen[itemID]; duplicate {
			return nil, ErrInvalid
		}
		seen[itemID] = struct{}{}
	}
	store, _, err := application.bind(ctx, scope, SnapshotAccess,
		OperationBatch, PermissionRead)
	if err != nil {
		return nil, err
	}
	result := make([]Snapshot, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		item, loadErr := store.Get(ctx, itemID)
		if loadErr != nil {
			return nil, mapStoreError(loadErr)
		}
		if item.ID != itemID || item.State == domain.ItemPurged {
			return nil, ErrNotFound
		}
		result = append(result, item)
	}
	return result, nil
}
