package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func schemaExecutorFinishFailedJob(
	ctx context.Context,
	transaction pgx.Tx,
	request schemaexecutor.FailureRequest,
	jobGeneration int64,
) (int64, error) {
	lease := request.Lease
	var generation int64
	var err error
	if request.State == schemaexecutor.FailureRetryable {
		if request.RetryDelay < 0 {
			return 0, fmt.Errorf("schema retry delay is negative")
		}
		err = transaction.QueryRow(ctx, schemaExecutorRetryJobSQL,
			lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
			lease.PlanSHA256(), jobGeneration,
			request.RetryDelay.Microseconds(),
		).Scan(&generation)
	} else {
		err = transaction.QueryRow(ctx, schemaExecutorFailJobSQL,
			lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
			lease.PlanSHA256(), jobGeneration,
			request.Code,
		).Scan(&generation)
	}
	if err != nil {
		return 0, schemaExecutorLeaseRow(err, "finish failed schema job")
	}
	return generation, nil
}
