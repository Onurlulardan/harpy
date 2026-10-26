package recoveryexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RejectRecoveryExecution(
	ctx context.Context,
	failure ports.RecoveryExecutionFailure,
) (bool, error) {
	var changed bool
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		return database.QueryRow(ctx, failRecoveryExecutionSQL,
			failure.RecoverySetID, failure.AttemptID,
			failure.LeaseToken, failure.Code).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("fail recovery execution: %w", err)
	}
	return changed, nil
}
