package recoveryexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) ClaimNextRecoverySet(
	ctx context.Context,
	attemptID string,
	leaseToken string,
) (ports.RecoveryExecutionClaim, bool, error) {
	var claim ports.RecoveryExecutionClaim
	found := true
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		err := database.QueryRow(
			ctx, claimNextRecoverySetSQL, attemptID, leaseToken,
		).Scan(
			&claim.AttemptID, &claim.LeaseToken, &claim.AttemptGeneration,
			&claim.RecoveryState, &claim.RecoverySetID,
			&claim.OperationID, &claim.EnvironmentID,
			&claim.Sequence, &claim.PrimaryBindingID, &claim.TenantSetSHA256,
			&claim.OwnerBindingID, &claim.OwnerBindingGeneration,
			&claim.DatabaseLineageID, &claim.DatabaseSystemID,
			&claim.SigningGeneration,
			&claim.RecoveryTrustBundleSHA256, &claim.ComponentInventory,
			&claim.DatabaseNow,
		)
		if store.rowNotFound(err) {
			found = false
			return nil
		}
		return err
	})
	if err != nil {
		return ports.RecoveryExecutionClaim{}, false,
			fmt.Errorf("claim next recovery set: %w", err)
	}
	return claim, found, nil
}
