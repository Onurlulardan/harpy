package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

const jobLeaseDuration = time.Duration(jobsv1.LeaseSeconds) * time.Second

func (repository *workerRepository) HeartbeatJob(
	ctx context.Context,
	claim JobClaim,
) (JobLease, error) {
	return withJobTransaction(
		ctx, repository, claim,
		func(transaction pgx.Tx) (JobLease, error) {
			arguments := append(
				authorityArguments(claim.authority),
				claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
			)
			var heartbeatAt, until time.Time
			var cancelRequested bool
			err := transaction.QueryRow(ctx, heartbeatJobSQL, arguments...).Scan(
				&heartbeatAt, &until, &cancelRequested,
			)
			if errors.Is(err, pgx.ErrNoRows) {
				return JobLease{}, ErrJobLeaseLost
			}
			if err != nil {
				return JobLease{}, fmt.Errorf("heartbeat job lease: %w", err)
			}
			return newJobLease(
				repository.identity, claim, heartbeatAt, until, cancelRequested,
			)
		},
	)
}
