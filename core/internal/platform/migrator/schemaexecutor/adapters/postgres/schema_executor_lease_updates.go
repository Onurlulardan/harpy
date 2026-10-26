package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func (store *schemaExecutorStore) MarkRunning(
	ctx context.Context,
	lease schemaexecutor.Lease,
) error {
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	return lease.Synchronize(func(snapshot schemaexecutor.LeaseSnapshot) (
		schemaexecutor.LeaseSnapshot, error,
	) {
		var generation int64
		err := store.pool.QueryRow(ctx, schemaExecutorMarkRunningSQL,
			lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
			lease.PlanSHA256(), snapshot.JobGeneration,
		).Scan(&generation)
		if err != nil {
			return snapshot, schemaExecutorLeaseRow(err, "mark schema job running")
		}
		snapshot.JobGeneration = generation
		return snapshot, nil
	})
}

func (store *schemaExecutorStore) Heartbeat(
	ctx context.Context,
	request schemaexecutor.HeartbeatRequest,
) error {
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	lease := request.Lease
	return lease.Synchronize(func(snapshot schemaexecutor.LeaseSnapshot) (
		schemaexecutor.LeaseSnapshot, error,
	) {
		var generation int64
		var expiresAt time.Time
		err := store.pool.QueryRow(ctx, schemaExecutorHeartbeatSQL,
			lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
			lease.PlanSHA256(), int(schemaexecutor.LeaseDuration.Seconds()),
			snapshot.JobGeneration,
		).Scan(&generation, &expiresAt)
		if err != nil {
			return snapshot, schemaExecutorLeaseRow(err, "heartbeat schema job")
		}
		snapshot.JobGeneration = generation
		snapshot.ExpiresAt = expiresAt
		return snapshot, nil
	})
}

type schemaExecutorProgress struct {
	tenantID           string
	checkpointRecordID *string
	processedRows      int64
}

func schemaExecutorLockLease(
	ctx context.Context,
	transaction pgx.Tx,
	lease schemaExecutorLeaseView,
	snapshot schemaexecutor.LeaseSnapshot,
) (schemaExecutorProgress, error) {
	var progress schemaExecutorProgress
	err := transaction.QueryRow(ctx, schemaExecutorLockLeaseSQL,
		lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
		lease.PlanSHA256(), lease.StepOrdinal(), lease.StepSHA256(),
		snapshot.JobGeneration, snapshot.StepGeneration, lease.StepAttempt(),
	).Scan(&progress.tenantID, &progress.checkpointRecordID, &progress.processedRows)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorProgress{}, schemaExecutorLeaseLost()
	}
	if err != nil {
		return schemaExecutorProgress{}, fmt.Errorf("lock schema lease: %w", err)
	}
	return progress, nil
}

func schemaExecutorLeaseRow(err error, operation string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorLeaseLost()
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func schemaExecutorLeaseLost() error {
	return schemaexecutor.NewFault(
		schemaexecutor.FaultLeaseLost, "schema_lease_lost", nil,
	)
}
