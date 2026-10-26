package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RecoveryKitClaim(
	ctx context.Context,
	operationID, actorUserID, sessionID string,
	occurredAt time.Time,
) (ports.RecoveryKitClaimRecord, error) {
	var record ports.RecoveryKitClaimRecord
	err := store.withBound(ctx, transactionCommand,
		func(database boundDatabase) error {
			err := database.QueryRow(ctx, claimRecoveryKitSQL,
				operationID, actorUserID, sessionID, occurredAt,
			).Scan(&record.ArtifactBytes, &record.ArtifactSHA256, &record.ExpiresAt)
			return store.mapDatabaseError(err)
		})
	if err != nil {
		return ports.RecoveryKitClaimRecord{}, fmt.Errorf("claim recovery kit: %w", err)
	}
	return record, nil
}
