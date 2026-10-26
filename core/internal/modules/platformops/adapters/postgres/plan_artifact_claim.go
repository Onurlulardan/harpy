package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) PlanArtifactClaim(
	ctx context.Context,
	claim ports.RecoveryArtifactClaim,
) (ports.RecoveryArtifactRecord, error) {
	var record ports.RecoveryArtifactRecord
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		err := database.QueryRow(ctx, claimRestorePlanSQL, claim.ArtifactID,
			claim.TenantID, claim.ActorID, claim.SessionID).
			Scan(&record.SHA256, &record.Bytes)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return ports.RecoveryArtifactRecord{}, fmt.Errorf("claim restore plan: %w", err)
	}
	return record, nil
}
