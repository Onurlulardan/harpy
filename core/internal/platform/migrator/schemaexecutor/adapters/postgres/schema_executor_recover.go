package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func (store *schemaExecutorStore) RecoverExpired(
	ctx context.Context,
) (bool, error) {
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	recovered := false
	err := store.withTransaction(ctx, func(transaction pgx.Tx) error {
		lease, found, err := schemaExecutorSelectExpired(ctx, transaction)
		if err != nil || !found {
			return err
		}
		if lease.StepAttempt() >= schemaexecutor.MaximumAttempts {
			if err := schemaExecutorExhaustLease(ctx, transaction, lease); err != nil {
				return err
			}
			recovered = true
			return nil
		}
		code := "lease_expired"
		detail := schemaExecutorDetailHash(code)
		finishedAt, err := schemaExecutorInsertAttempt(
			ctx, transaction, lease, "lease_expired", &code, &detail,
		)
		if err != nil {
			return fmt.Errorf("record expired schema lease: %w", err)
		}
		snapshot := schemaexecutor.LeaseSnapshot{
			JobGeneration: lease.jobGeneration, StepGeneration: lease.stepGeneration,
			ExpiresAt: lease.leaseExpiresAt,
		}
		stepGeneration, err := schemaExecutorFinishStep(
			ctx, transaction, lease, snapshot,
			"lease_expired", "migrator", &code, finishedAt,
		)
		if err != nil {
			return err
		}
		snapshot.StepGeneration = stepGeneration
		if err := schemaExecutorExpireAndQueue(
			ctx, transaction, lease, snapshot,
		); err != nil {
			return err
		}
		recovered = true
		return nil
	})
	return recovered, err
}

func schemaExecutorSelectExpired(
	ctx context.Context,
	transaction pgx.Tx,
) (schemaExecutorRecoveredLease, bool, error) {
	var lease schemaExecutorRecoveredLease
	err := transaction.QueryRow(ctx, schemaExecutorSelectExpiredSQL).Scan(
		&lease.jobID, &lease.tenantID, &lease.leaseID, &lease.executorID,
		&lease.jobAttempt, &lease.stepAttempt, &lease.stepOrdinal, &lease.planSHA256,
		&lease.leaseExpiresAt, &lease.stepSHA256, &lease.stepStartedAt,
		&lease.jobGeneration, &lease.stepGeneration,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaExecutorRecoveredLease{}, false, nil
	}
	if err != nil {
		return schemaExecutorRecoveredLease{}, false, err
	}
	return lease, true, nil
}

func schemaExecutorExhaustLease(
	ctx context.Context,
	transaction pgx.Tx,
	lease schemaExecutorRecoveredLease,
) error {
	code := "attempt_limit_exhausted"
	detail := schemaExecutorDetailHash(code)
	finishedAt, err := schemaExecutorInsertAttempt(
		ctx, transaction, lease, "failed", &code, &detail,
	)
	if err != nil {
		return err
	}
	snapshot := schemaexecutor.LeaseSnapshot{
		JobGeneration: lease.jobGeneration, StepGeneration: lease.stepGeneration,
		ExpiresAt: lease.leaseExpiresAt,
	}
	stepGeneration, err := schemaExecutorFinishStep(
		ctx, transaction, lease, snapshot, "failed", "none", &code, finishedAt,
	)
	if err != nil {
		return err
	}
	snapshot.StepGeneration = stepGeneration
	var jobGeneration int64
	err = transaction.QueryRow(ctx, schemaExecutorFailExpiredJobSQL,
		lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
		lease.PlanSHA256(), snapshot.JobGeneration, code,
	).Scan(&jobGeneration)
	return schemaExecutorLeaseRow(err, "fail exhausted schema job")
}

func schemaExecutorDetailHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
