package maintenance

import (
	"context"
	"fmt"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func (store *store) executeDirectoryTombstonePurge(
	ctx context.Context, claim maintenanceapp.Claim, nextID uuid.UUID,
) (maintenanceapp.Outcome, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || !store.runtime.usable() {
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
	values, err := claimValues(claim)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	if receipt, found, err := store.readReceipt(ctx, claim); found || err != nil {
		return receipt, err
	}
	rows, err := store.database.Query(ctx, readDirectoryTombstoneCandidatesSQL)
	if err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("read directory purge candidates: %w", err)
	}
	defer rows.Close()
	var purged int64
	for rows.Next() {
		var identifierID uuid.UUID
		var revision int64
		if err := rows.Scan(&identifierID, &revision); err != nil {
			return maintenanceapp.Outcome{}, fmt.Errorf("scan directory purge candidate: %w", err)
		}
		if identifierID == uuid.Nil || revision < 1 {
			return maintenanceapp.Outcome{}, ErrRuntimeInvalid
		}
		var affected int64
		if err := store.database.QueryRow(
			ctx, purgeDirectoryProjectionSQL, claim.TenantID, identifierID, revision,
		).Scan(&affected); err != nil {
			return maintenanceapp.Outcome{}, fmt.Errorf("purge directory projection: %w", err)
		}
		if affected < 0 {
			return maintenanceapp.Outcome{}, ErrRuntimeInvalid
		}
		purged += affected
	}
	if err := rows.Err(); err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("iterate directory purge candidates: %w", err)
	}
	finish := append(values, purged, purged, nextID)
	outcome, err := scanOutcome(
		store.database.QueryRow(ctx, executeDirectoryTombstonePurgeSQL, finish...), claim.Kind,
	)
	if err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("finish directory tombstone purge: %w", err)
	}
	return outcome, nil
}
