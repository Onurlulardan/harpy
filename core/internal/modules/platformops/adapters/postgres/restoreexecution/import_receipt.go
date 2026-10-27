package restoreexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) ImportRestoreReceipt(
	ctx context.Context,
	record ports.RestoreReceiptImport,
) (bool, error) {
	var imported bool
	err := store.withBound(ctx, func(database boundDatabase) error {
		return database.QueryRow(
			ctx, importReceiptSQL, record.OperationID, record.RecoverySetID,
			record.EnvironmentID, record.PlanSHA256, record.ManifestSHA256,
			record.ReceiptSHA256, record.ReceiptBytes,
			nullableUUID(record.PreviousDatabaseSystemID),
			record.NewDatabaseSystemID, record.ObjectMerkleSHA256,
			record.LocatorRemapSHA256, record.SafetyEvidenceSHA256,
			record.DrainAckSHA256, record.SwitchAckSHA256,
			record.ControlDatabaseLocatorSHA256,
			record.ScratchDatabaseLocatorSHA256,
			string(record.ComponentInventory),
		).Scan(&imported)
	})
	if err != nil {
		return false, fmt.Errorf("import restore receipt: %w", err)
	}
	return imported, nil
}

func nullableUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}
