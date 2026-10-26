package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) ResolveRestore(
	ctx context.Context,
	operationID string,
) (ports.RestoreRecord, error) {
	var record ports.RestoreRecord
	var disposition string
	var receiptSHA, databaseSystemID *string
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		err := database.QueryRow(ctx, getRestoreSQL, operationID).Scan(
			&record.OperationID, &record.RecoverySetID, &record.EnvironmentID,
			&disposition, &record.State, &record.PlanSHA256,
			&receiptSHA, &databaseSystemID, &record.CreatedAt,
			&record.UpdatedAt, &record.CompletedAt)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.RestoreRecord{}, fmt.Errorf("get restore: %w", err)
	}
	record.Disposition = recoveryset.TargetDisposition(disposition)
	if receiptSHA != nil {
		record.ReceiptSHA256 = *receiptSHA
	}
	if databaseSystemID != nil {
		record.NewDatabaseSystemID = *databaseSystemID
	}
	return record, nil
}
