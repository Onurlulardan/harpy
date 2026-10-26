package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (store *Store) ReleaseHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
	digest [32]byte,
) (bool, error) {
	var changed bool
	err := store.with(ctx, transactionCommand,
		func(database recoveryDatabase) error {
			return database.QueryRow(
				ctx, releaseHoldSQL, recoverySetID, digest[:],
			).Scan(&changed)
		})
	return changed, err
}
