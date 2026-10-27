package postrestore

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (store *Store) FailChallenge(
	ctx context.Context, operationID, environmentID, restoreOperationID uuid.UUID,
	occurredAt time.Time,
) error {
	return store.global(ctx, transactionCommand, func(database database) error {
		var applied bool
		return database.QueryRow(ctx, failChallengeSQL, operationID, environmentID,
			restoreOperationID, occurredAt).Scan(&applied)
	})
}
