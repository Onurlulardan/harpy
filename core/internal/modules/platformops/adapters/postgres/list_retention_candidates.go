package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) ListRetentionCandidates(
	ctx context.Context,
	environmentID string,
) ([]ports.RetentionRecord, error) {
	result := make([]ports.RetentionRecord, 0)
	err := store.withBound(ctx, transactionSnapshotQuery, func(database boundDatabase) error {
		rows, err := database.Query(ctx, listRetentionCandidatesSQL, environmentID)
		if err != nil {
			return store.mapDatabaseError(err)
		}
		defer rows.Close()
		for rows.Next() {
			var record ports.RetentionRecord
			if err = rows.Scan(&record.RecoverySetID, &record.ManifestSHA256,
				&record.ProtectedAt, &record.LegalHold); err != nil {
				return store.mapDatabaseError(err)
			}
			result = append(result, record)
		}
		return store.mapDatabaseError(rows.Err())
	})
	if err != nil {
		return nil, fmt.Errorf("list recovery retention candidates: %w", err)
	}
	return result, nil
}
