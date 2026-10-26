package recoveryexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RecordRecoveryComponent(
	ctx context.Context,
	record ports.RecoveryComponentRecord,
) (bool, error) {
	var changed bool
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		return database.QueryRow(
			ctx, recordRecoveryComponentSQL,
			record.RecoverySetID, record.AttemptID, record.LeaseToken,
			record.Kind, record.ObjectKey, record.ObjectVersion,
			record.CipherSHA256, record.CiphertextBytes,
		).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("record recovery component: %w", err)
	}
	return changed, nil
}
