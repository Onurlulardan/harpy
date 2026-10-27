package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RecoveryControlLookup(
	ctx context.Context,
	operationID string,
) (ports.RecoveryControlOperationRecord, error) {
	var record ports.RecoveryControlOperationRecord
	var digest, failure *string
	var expires *time.Time
	err := store.withBound(ctx, transactionSnapshotQuery,
		func(database boundDatabase) error {
			err := database.QueryRow(ctx, getRecoveryControlOperationSQL, operationID).Scan(
				&record.OperationID, &record.EnvironmentID, &record.Kind,
				&record.State, &record.KitClaimState, &digest, &expires,
				&failure, &record.CreatedAt)
			return store.mapDatabaseError(err)
		})
	if err != nil {
		return ports.RecoveryControlOperationRecord{},
			fmt.Errorf("get recovery control operation: %w", err)
	}
	if digest != nil {
		record.ArtifactSHA256 = *digest
	}
	if failure != nil {
		record.FailureCode = *failure
	}
	record.ExpiresAt = expires
	return record, nil
}
