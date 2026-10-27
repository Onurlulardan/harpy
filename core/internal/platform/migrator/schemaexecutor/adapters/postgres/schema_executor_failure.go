package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func (store *schemaExecutorStore) RecordFailure(
	ctx context.Context,
	request schemaexecutor.FailureRequest,
) error {
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	transaction, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer store.rollbackTransaction(transaction)
	return request.Lease.Synchronize(func(snapshot schemaexecutor.LeaseSnapshot) (
		schemaexecutor.LeaseSnapshot, error,
	) {
		updated, err := store.recordFailureTransaction(
			ctx, transaction, request, snapshot,
		)
		if err != nil {
			return snapshot, err
		}
		if err := transaction.Commit(ctx); err != nil {
			return snapshot, err
		}
		return updated, nil
	})
}

func (store *schemaExecutorStore) recordFailureTransaction(
	ctx context.Context,
	transaction pgx.Tx,
	request schemaexecutor.FailureRequest,
	snapshot schemaexecutor.LeaseSnapshot,
) (schemaexecutor.LeaseSnapshot, error) {
	lease := request.Lease
	if _, err := schemaExecutorLockLease(
		ctx, transaction, lease, snapshot,
	); err != nil {
		return snapshot, err
	}
	state, nextExecutor := string(request.State), "none"
	if request.State == schemaexecutor.FailureRetryable {
		nextExecutor = "migrator"
	}
	code, detail := request.Code, request.RedactedDetailSHA256
	finishedAt, err := schemaExecutorInsertAttempt(
		ctx, transaction, lease, state, &code, &detail,
	)
	if err != nil {
		return snapshot, fmt.Errorf("record schema step failure: %w", err)
	}
	stepGeneration, err := schemaExecutorFinishStep(
		ctx, transaction, lease, snapshot, state,
		nextExecutor, &code, finishedAt,
	)
	if err != nil {
		return snapshot, err
	}
	snapshot.StepGeneration = stepGeneration
	jobGeneration, err := schemaExecutorFinishFailedJob(
		ctx, transaction, request, snapshot.JobGeneration,
	)
	if err != nil {
		return snapshot, err
	}
	snapshot.JobGeneration = jobGeneration
	return snapshot, nil
}
