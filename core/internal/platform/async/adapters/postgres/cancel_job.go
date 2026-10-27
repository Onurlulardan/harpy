package postgres

import (
	"context"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) CancelJob(
	ctx context.Context,
	claim JobClaim,
	cancellation JobCancellation,
) (time.Time, error) {
	if !cancellation.valid() {
		return time.Time{}, ErrJobFailureInvalid
	}
	failure, err := NewJobFailure(
		jobsv1.ErrorCancelled, cancellation.code, cancellation.digest, 0,
	)
	if err != nil {
		return time.Time{}, err
	}
	decision, err := decideJobFailure(claim, failure)
	if err != nil {
		return time.Time{}, err
	}
	result, err := withJobTransaction(
		ctx, repository, claim,
		func(transaction pgx.Tx) (JobFailureResult, error) {
			locked, err := lockJobClaim(ctx, transaction, claim)
			if err != nil {
				return JobFailureResult{}, err
			}
			if locked.effectStarted.Valid {
				return JobFailureResult{}, ErrExternalEffectPermit
			}
			occurredAt, err := failJobAttempt(
				ctx, transaction, claim, failure, decision, nil, nil,
			)
			if err != nil {
				return JobFailureResult{}, err
			}
			return persistFailedJob(
				ctx, transaction, claim, failure, decision,
				occurredAt, time.Time{},
			)
		},
	)
	return result.occurredAt, err
}
