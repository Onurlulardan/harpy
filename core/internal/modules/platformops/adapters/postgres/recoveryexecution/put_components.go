package recoveryexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) PutRecoveryComponents(
	ctx context.Context,
	record ports.RecoveryComponentBatch,
) (bool, error) {
	var changed bool
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		return database.QueryRow(
			ctx, putRecoveryComponentsSQL,
			record.RecoverySetID, record.AttemptID,
			record.LeaseToken, []byte(record.Components),
		).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("record recovery component batch: %w", err)
	}
	return changed, nil
}
