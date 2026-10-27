package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func schemaExecutorFinishStep(
	ctx context.Context,
	transaction pgx.Tx,
	lease schemaExecutorLeaseView,
	snapshot schemaexecutor.LeaseSnapshot,
	state string,
	nextExecutor string,
	errorCode *string,
	finishedAt time.Time,
) (int64, error) {
	var generation int64
	err := transaction.QueryRow(ctx, schemaExecutorFinishStepSQL,
		lease.JobID(), lease.StepOrdinal(), lease.StepAttempt(),
		lease.PlanSHA256(), lease.StepSHA256(), snapshot.StepGeneration,
		state, errorCode, nextExecutor, finishedAt,
	).Scan(&generation)
	if err != nil {
		return 0, schemaExecutorLeaseRow(err, "finish schema step")
	}
	return generation, nil
}
