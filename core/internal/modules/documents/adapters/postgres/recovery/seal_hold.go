package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (store *Store) SealHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
	digest [32]byte,
	count int64,
) (bool, error) {
	var changed bool
	err := store.with(ctx, transactionCommand,
		func(database recoveryDatabase) error {
			return database.QueryRow(
				ctx, sealHoldSQL, recoverySetID, digest[:], count,
			).Scan(&changed)
		})
	return changed, err
}
