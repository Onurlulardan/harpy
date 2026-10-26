package restoreexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) ClaimRestoreExecution(
	ctx context.Context,
	identity ports.RestoreExecutionIdentity,
	planSHA256 string,
) (ports.RestoreExecutionClaim, bool, error) {
	var claim ports.RestoreExecutionClaim
	found := true
	err := store.withBound(ctx, func(database boundDatabase) error {
		err := database.QueryRow(
			ctx, claimSQL, identity.OperationID, identity.AttemptID,
			identity.LeaseToken, planSHA256,
		).Scan(
			&claim.AttemptID, &claim.LeaseToken, &claim.AttemptGeneration,
			&claim.RestoreState, &claim.OperationID, &claim.RecoverySetID,
			&claim.EnvironmentID, &claim.TargetDisposition,
			&claim.PlanSHA256, &claim.PlanBytes,
			&claim.SafetyRecoverySetID, &claim.SafetyRecoverySetSHA256,
			&claim.WaiverReason, &claim.ProbeSHA256,
			&claim.DestructiveConfirmed, &claim.ManifestSHA256,
			&claim.ManifestBytes, &claim.PreviousDatabaseSystemID,
			&claim.ComponentInventory,
			&claim.ControlDatabaseLocatorSHA256,
			&claim.ScratchDatabaseLocatorSHA256,
			&claim.DrainAckSHA256, &claim.SwitchAckSHA256,
			&claim.ArchiveReadbackSHA256,
			&claim.NewDatabaseSystemID,
			&claim.ObjectMerkleSHA256,
			&claim.LocatorRemapSHA256,
			&claim.SafetyEvidenceSHA256,
			&claim.DatabaseNow,
		)
		if store.rowNotFound(err) {
			found = false
			return nil
		}
		return err
	})
	if err != nil {
		return ports.RestoreExecutionClaim{}, false,
			fmt.Errorf("claim restore execution: %w", err)
	}
	return claim, found, nil
}
