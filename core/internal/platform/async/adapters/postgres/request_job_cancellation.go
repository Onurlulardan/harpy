package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) RequestJobCancellation(
	ctx context.Context,
	control JobControl,
	cancellation JobCancellation,
) (JobCancelResult, error) {
	if !control.valid(repository.identity) || !cancellation.valid() {
		return JobCancelResult{}, ErrJobControlInvalid
	}
	return withJobAuthorityTransaction(
		ctx, repository, control.authority,
		func(transaction pgx.Tx) (JobCancelResult, error) {
			arguments := append(
				authorityArguments(control.authority),
				control.jobID, cancellation.code, cancellation.digest,
			)
			var state string
			var requestedAt time.Time
			err := transaction.QueryRow(
				ctx, requestJobCancellationSQL, arguments...,
			).Scan(&state, &requestedAt)
			if errors.Is(err, pgx.ErrNoRows) {
				return JobCancelResult{}, ErrJobNotControllable
			}
			if err != nil {
				return JobCancelResult{}, fmt.Errorf("request job cancellation: %w", err)
			}
			result := JobCancelResult{
				state: jobsv1.State(state), requestedAt: requestedAt,
			}
			if result.state != jobsv1.StateRunning &&
				result.state != jobsv1.StateCancelled {
				return JobCancelResult{}, ErrJobNotControllable
			}
			return result, nil
		},
	)
}
