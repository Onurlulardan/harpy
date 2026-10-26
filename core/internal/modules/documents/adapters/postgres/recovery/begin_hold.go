package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (store *Store) BeginHold(
	ctx context.Context,
	recoverySetID uuid.UUID,
) (string, error) {
	var state string
	err := store.with(ctx, transactionCommand,
		func(database recoveryDatabase) error {
			return database.QueryRow(ctx, beginHoldSQL,
				recoverySetID, store.environmentID, store.tenantID,
				store.tenantSetSHA256[:]).Scan(&state)
		})
	return state, err
}
