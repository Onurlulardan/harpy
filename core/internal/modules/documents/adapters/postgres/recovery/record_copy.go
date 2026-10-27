package recovery

import (
	"context"

	recoveryapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/recovery"
)

func (store *Store) RecordCopy(
	ctx context.Context,
	receipt recoveryapp.CopyReceipt,
) error {
	return store.with(ctx, transactionCommand,
		func(database recoveryDatabase) error {
			var ignored any
			return database.QueryRow(ctx, recordCopySQL,
				receipt.RecoverySetID, receipt.Item.VersionID,
				receipt.Item.ObjectKey, receipt.Item.ObjectVersion,
				receipt.TargetObjectKey, receipt.TargetObjectVersion,
				receipt.Item.CiphertextSHA256[:], receipt.Item.CiphertextBytes,
				receipt.Item.KeyID).Scan(&ignored)
		})
}
