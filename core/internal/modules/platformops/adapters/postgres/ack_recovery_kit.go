package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) AcknowledgeRecoveryKit(
	ctx context.Context,
	operationID, actorUserID, sessionID string,
	artifactSHA256, ackProof []byte,
	occurredAt time.Time,
) error {
	var acknowledged bool
	err := store.withBound(ctx, transactionCommand,
		func(database boundDatabase) error {
			err := database.QueryRow(ctx, ackRecoveryKitSQL,
				operationID, actorUserID, sessionID, artifactSHA256,
				ackProof, occurredAt,
			).Scan(&acknowledged)
			return store.mapDatabaseError(err)
		})
	if err != nil {
		return fmt.Errorf("acknowledge recovery kit: %w", err)
	}
	if !acknowledged {
		return ports.ErrConflict
	}
	return nil
}
