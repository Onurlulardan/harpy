package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func schemaExecutorCompleteStep(
	ctx context.Context,
	transaction pgx.Tx,
	request schemaexecutor.ApplyRequest,
	snapshot schemaexecutor.LeaseSnapshot,
	state string,
) (schemaexecutor.LeaseSnapshot, error) {
	lease := request.Lease
	finishedAt, err := schemaExecutorInsertAttempt(
		ctx, transaction, lease, state, nil, nil,
	)
	if err != nil {
		return snapshot, fmt.Errorf("record schema step success: %w", err)
	}
	stepGeneration, err := schemaExecutorFinishStep(
		ctx, transaction, lease, snapshot, state, "none", nil, finishedAt,
	)
	if err != nil {
		return snapshot, err
	}
	snapshot.StepGeneration = stepGeneration
	nextOrdinal := lease.StepOrdinal() + 1
	if nextOrdinal == request.Plan.StepCount() {
		var jobGeneration int64
		err := transaction.QueryRow(ctx, schemaExecutorCompleteJobSQL,
			lease.JobID(), lease.LeaseID(), lease.ExecutorID(),
			lease.JobAttempt(), lease.PlanSHA256(), snapshot.JobGeneration,
		).Scan(&jobGeneration)
		if err != nil {
			return snapshot, schemaExecutorLeaseRow(err, "complete schema job")
		}
		snapshot.JobGeneration = jobGeneration
		return snapshot, nil
	}
	next, exists := request.Plan.StepAt(nextOrdinal)
	if !exists {
		return snapshot, fmt.Errorf("next schema step is missing")
	}
	jobState := "queued"
	if next.ExecutorKind() == schemav1.FormsRuntimeExecutor {
		jobState = "awaiting_runtime"
	}
	var jobGeneration int64
	err = transaction.QueryRow(ctx, schemaExecutorAdvanceJobSQL,
		lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
		lease.PlanSHA256(), snapshot.JobGeneration,
		jobState, string(next.ExecutorKind()),
	).Scan(&jobGeneration)
	if err != nil {
		return snapshot, schemaExecutorLeaseRow(err, "advance schema job")
	}
	snapshot.JobGeneration = jobGeneration
	return snapshot, nil
}
