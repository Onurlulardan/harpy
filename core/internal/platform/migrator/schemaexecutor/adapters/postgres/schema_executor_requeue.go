package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func schemaExecutorExpireAndQueue(
	ctx context.Context,
	transaction pgx.Tx,
	lease schemaExecutorLeaseView,
	snapshot schemaexecutor.LeaseSnapshot,
) error {
	var jobGeneration int64
	err := transaction.QueryRow(ctx, schemaExecutorExpireJobSQL,
		lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
		lease.PlanSHA256(), snapshot.JobGeneration,
	).Scan(&jobGeneration)
	if err := schemaExecutorLeaseRow(err, "expire schema job"); err != nil {
		return err
	}
	err = transaction.QueryRow(ctx, schemaExecutorRecoverJobSQL,
		lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
		lease.PlanSHA256(), jobGeneration,
	).Scan(&jobGeneration)
	return schemaExecutorLeaseRow(err, "queue expired schema job")
}
