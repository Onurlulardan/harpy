package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (store *Store) FailHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
	code string,
) (bool, error) {
	var changed bool
	err := store.with(ctx, transactionCommand,
		func(database recoveryDatabase) error {
			return database.QueryRow(ctx, failHoldSQL, recoverySetID, code).Scan(&changed)
		})
	return changed, err
}
