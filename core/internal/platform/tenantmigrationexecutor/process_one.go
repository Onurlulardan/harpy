package tenantmigrationexecutor

import (
	"context"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/google/uuid"
)

func (executor *Executor) processOne(
	ctx context.Context, permit admission.Permit, profileID uuid.UUID,
) (bool, error) {
	if !permit.Granted() {
		return false, fmt.Errorf("tenant migration admission permit is invalid")
	}
	leaseID, runID, expiryID, err := executor.claimIDs()
	if err != nil {
		return false, err
	}
	claim, found, err := executor.store.Claim(
		ctx, profileID, leaseID, runID, expiryID, executor.leaseTTL,
	)
	if err != nil || !found {
		return false, err
	}
	workContext, cancel := context.WithCancel(ctx)
	heartbeatDone := make(chan error, 1)
	go executor.heartbeatClaim(workContext, claim, cancel, heartbeatDone)
	receipt, migrationErr := executor.migrator.Migrate(workContext, claim)
	cancel()
	heartbeatErr := <-heartbeatDone
	if heartbeatErr != nil {
		return true, heartbeatErr
	}
	if migrationErr != nil {
		return true, executor.recordFailure(context.WithoutCancel(ctx), claim, migrationErr)
	}
	receiptID, err := executor.newID()
	if err != nil {
		return true, err
	}
	if !validDigest(receipt.ReceiptSHA256) || !validDigest(receipt.EvidenceSHA256) {
		return true, fmt.Errorf("tenant migration receipt is invalid")
	}
	return true, executor.store.Complete(
		context.WithoutCancel(ctx), claim, receiptID, receipt,
	)
}

func (executor *Executor) claimIDs() (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	values := make([]uuid.UUID, 3)
	for index := range values {
		value, err := executor.newID()
		if err != nil {
			return uuid.Nil, uuid.Nil, uuid.Nil, err
		}
		values[index] = value
	}
	return values[0], values[1], values[2], nil
}

func failureDetails(err error) (string, string) {
	switch {
	case errors.Is(err, ErrAmbiguousOutcome):
		return "reconcile_required", "migration_outcome_ambiguous"
	case errors.Is(err, ErrCleanupRequired):
		return "cleanup_required", "orphan_schema_requires_cleanup"
	default:
		return "failed", "migration_execution_failed"
	}
}
