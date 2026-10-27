package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

var errSchemaExecutorNoClaim = errors.New("no claimable schema job")

func (store *schemaExecutorStore) Claim(
	ctx context.Context,
	request schemaexecutor.ClaimRequest,
) (schemaexecutor.ClaimRecord, bool, error) {
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	var claim schemaexecutor.ClaimRecord
	err := store.withTransaction(ctx, func(transaction pgx.Tx) error {
		job, stepGeneration, err := schemaExecutorSelectDueJob(ctx, transaction)
		if err != nil {
			return err
		}
		claim.Job = job
		var expiresAt time.Time
		err = transaction.QueryRow(ctx, schemaExecutorLeaseJobSQL,
			job.JobID, request.LeaseID, request.ExecutorID,
			int(schemaexecutor.LeaseDuration.Seconds()), job.Generation,
		).Scan(&claim.Job.Attempt, &expiresAt, &claim.Job.Generation)
		if err != nil {
			return fmt.Errorf("lease schema job: %w", err)
		}
		var stepAttempt int
		var stepHash string
		var startedAt time.Time
		var startedStepGeneration int64
		err = transaction.QueryRow(ctx, schemaExecutorStartStepSQL,
			job.JobID, job.NextStepOrdinal, stepGeneration,
		).Scan(&stepAttempt, &stepHash, &startedAt, &startedStepGeneration)
		if err != nil {
			return fmt.Errorf("start schema step: %w", err)
		}
		claim.Steps, err = schemaExecutorReadSteps(ctx, transaction, job.JobID)
		if err != nil {
			return err
		}
		claim.Lease = schemaexecutor.LeaseRecord{
			JobID: job.JobID, TenantID: job.TenantID,
			LeaseID: request.LeaseID, ExecutorID: request.ExecutorID,
			JobAttempt: claim.Job.Attempt, StepAttempt: stepAttempt,
			StepOrdinal: job.NextStepOrdinal,
			PlanSHA256:  job.PlanSHA256, StepSHA256: stepHash,
			StepStartedAt: startedAt, LeaseExpiresAt: expiresAt,
			JobGeneration: claim.Job.Generation, StepGeneration: startedStepGeneration,
		}
		return nil
	})
	if errors.Is(err, errSchemaExecutorNoClaim) {
		return schemaexecutor.ClaimRecord{}, false, nil
	}
	return claim, err == nil, err
}

func schemaExecutorSelectDueJob(
	ctx context.Context,
	transaction pgx.Tx,
) (schemaexecutor.JobRecord, int64, error) {
	var job schemaexecutor.JobRecord
	var canonical string
	var stepGeneration int64
	err := transaction.QueryRow(ctx, schemaExecutorSelectDueJobSQL).Scan(
		&job.JobID, &job.TenantID, &job.Owner, &job.ResourceKind,
		&job.ResourceID, &job.ResourceRevision, &job.CompilerVersion,
		&job.PlanContractVersion, &canonical, &job.PlanSHA256,
		&job.StepCount, &job.NextStepOrdinal, &job.Attempt,
		&job.Generation, &stepGeneration,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return schemaexecutor.JobRecord{}, 0, errSchemaExecutorNoClaim
	}
	if err != nil {
		return schemaexecutor.JobRecord{}, 0, fmt.Errorf("select due schema job: %w", err)
	}
	job.CanonicalPlan = []byte(canonical)
	return job, stepGeneration, nil
}
