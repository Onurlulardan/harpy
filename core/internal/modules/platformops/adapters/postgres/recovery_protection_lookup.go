package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RecoveryProtectionLookup(
	ctx context.Context,
	environmentID string,
) (ports.RecoveryProtectionRecord, error) {
	var record ports.RecoveryProtectionRecord
	err := store.withBound(ctx, transactionSnapshotQuery,
		func(database boundDatabase) error {
			err := database.QueryRow(ctx, getRecoveryProtectionSQL, environmentID).Scan(
				&record.EnvironmentID, &record.State, &record.Generation,
				&record.KitAcknowledged, &record.UpdatedAt)
			return store.mapDatabaseError(err)
		})
	if err != nil {
		return ports.RecoveryProtectionRecord{}, fmt.Errorf("get recovery protection: %w", err)
	}
	return record, nil
}
