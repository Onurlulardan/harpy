package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RecoverySetListing(
	ctx context.Context,
	environmentID string,
) ([]ports.RecoverySetRecord, error) {
	result := make([]ports.RecoverySetRecord, 0)
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		rows, err := database.Query(ctx, listRecoverySetsSQL, environmentID)
		if err != nil {
			return store.mapDatabaseError(err)
		}
		defer rows.Close()
		for rows.Next() {
			var record ports.RecoverySetRecord
			var manifestSHA, databaseSystemID *string
			if err = rows.Scan(&record.RecoverySetID, &record.OperationID,
				&record.EnvironmentID, &record.Sequence, &record.State, &manifestSHA,
				&record.LegalHold, &databaseSystemID, &record.ProtectedAt); err != nil {
				return store.mapDatabaseError(err)
			}
			if manifestSHA != nil {
				record.ManifestSHA256 = *manifestSHA
			}
			if databaseSystemID != nil {
				record.DatabaseSystemID = *databaseSystemID
			}
			result = append(result, record)
		}
		return store.mapDatabaseError(rows.Err())
	})
	if err != nil {
		return nil, fmt.Errorf("list recovery sets: %w", err)
	}
	return result, nil
}
