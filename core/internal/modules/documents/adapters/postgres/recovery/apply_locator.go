package recovery

import (
	"context"

	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/recovery"
)

func (store *Store) ApplyLocator(
	ctx context.Context,
	remap recoveryapp.LocatorRemap,
) (bool, error) {
	var changed bool
	err := store.with(ctx, transactionCommand,
		func(database recoveryDatabase) error {
			return database.QueryRow(ctx, applyLocatorSQL,
				remap.RestoreOperationID, remap.RecoverySetID, remap.Item.VersionID,
				remap.Item.ObjectKey, remap.Item.ObjectVersion, remap.TargetObjectKey,
				remap.TargetObjectVersion, remap.TargetObjectGeneration,
				remap.Item.CiphertextSHA256[:]).Scan(&changed)
		})
	return changed, err
}
