package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
	"github.com/google/uuid"
)

func (store *Store) LoadSigningBundle(
	ctx context.Context,
	environmentID string,
	generation int64,
) (ports.SigningBundleRecord, error) {
	if _, err := uuid.Parse(environmentID); err != nil || generation <= 0 {
		return ports.SigningBundleRecord{}, ports.ErrNotFound
	}
	var record ports.SigningBundleRecord
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		err := database.QueryRow(ctx, getSigningBundleSQL,
			environmentID, generation).Scan(
			&record.EnvironmentID, &record.Generation, &record.SHA256,
			&record.BundleBytes, &record.CreatedAt)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.SigningBundleRecord{}, fmt.Errorf("load recovery signing bundle: %w", err)
	}
	return record, nil
}
