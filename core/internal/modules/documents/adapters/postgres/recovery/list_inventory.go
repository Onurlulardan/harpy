package recovery

import (
	"context"

	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/recovery"
	"github.com/google/uuid"
)

func (store *Store) ListInventory(
	ctx context.Context,
	recoverySetID uuid.UUID,
) ([]recoveryapp.InventoryItem, error) {
	items := make([]recoveryapp.InventoryItem, 0)
	err := store.with(ctx, transactionSnapshotQuery,
		func(database recoveryDatabase) error {
			rows, err := database.Query(ctx, inventorySQL, recoverySetID, store.tenantID)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var item recoveryapp.InventoryItem
				var digest []byte
				if err = rows.Scan(&item.VersionID, &item.ObjectKey, &item.ObjectVersion,
					&item.ObjectGeneration, &digest, &item.CiphertextBytes,
					&item.KeyID); err != nil {
					return err
				}
				if len(digest) != 32 {
					return recoveryapp.ErrUnavailable
				}
				copy(item.CiphertextSHA256[:], digest)
				items = append(items, item)
			}
			return rows.Err()
		})
	return items, err
}
