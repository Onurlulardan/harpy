package recoveryexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) UpdateRecoveryLease(
	ctx context.Context,
	identity ports.RecoveryExecutionIdentity,
) (bool, error) {
	return store.executionChange(ctx, heartbeatRecoveryExecutionSQL, identity)
}

func (store *Store) executionChange(
	ctx context.Context,
	statement string,
	identity ports.RecoveryExecutionIdentity,
) (bool, error) {
	var changed bool
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		return database.QueryRow(ctx, statement, identity.RecoverySetID,
			identity.AttemptID, identity.LeaseToken).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("change recovery execution: %w", err)
	}
	return changed, nil
}
