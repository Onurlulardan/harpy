package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) GetRecoverySet(
	ctx context.Context,
	recoverySetID string,
) (ports.RecoverySetRecord, error) {
	var record ports.RecoverySetRecord
	var manifestSHA, databaseSystemID *string
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		err := database.QueryRow(ctx, getRecoverySetSQL, recoverySetID).Scan(
			&record.RecoverySetID, &record.OperationID, &record.EnvironmentID,
			&record.Sequence, &record.State, &manifestSHA, &record.ManifestBytes,
			&record.LegalHold, &databaseSystemID, &record.RecoverySigningGeneration,
			&record.RecoveryTrustBundleSHA256, &record.ProtectedAt,
			&record.CreatedAt, &record.UpdatedAt)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.RecoverySetRecord{}, fmt.Errorf("get recovery set: %w", err)
	}
	if manifestSHA != nil {
		record.ManifestSHA256 = *manifestSHA
	}
	if databaseSystemID != nil {
		record.DatabaseSystemID = *databaseSystemID
	}
	return record, nil
}
