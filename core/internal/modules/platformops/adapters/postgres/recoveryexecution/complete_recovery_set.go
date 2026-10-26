package recoveryexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) CompleteRecoverySet(
	ctx context.Context,
	record ports.RecoveryCompletionRecord,
) (bool, error) {
	var changed bool
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		return database.QueryRow(
			ctx, completeRecoverySetSQL,
			record.RecoverySetID, record.AttemptID, record.LeaseToken,
			record.ManifestSHA256, record.ManifestBytes, record.DatabaseSystemID,
		).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("complete recovery set: %w", err)
	}
	return changed, nil
}
