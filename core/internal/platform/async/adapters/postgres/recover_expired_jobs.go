package postgres

import (
	"context"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) RecoverExpiredJobs(
	ctx context.Context,
	authority JobAuthority,
	limit int,
) (JobRecoveryResult, error) {
	if limit < 1 || limit > jobsv1.MaximumClaimBatch {
		return JobRecoveryResult{}, ErrJobRecoveryLimit
	}
	return withJobAuthorityTransaction(
		ctx, repository, authority,
		func(transaction pgx.Tx) (JobRecoveryResult, error) {
			rows, err := transaction.Query(
				ctx, lockExpiredJobsSQL,
				append(authorityArguments(authority), limit, string(authority.mode))...,
			)
			if err != nil {
				return JobRecoveryResult{}, fmt.Errorf("lock expired jobs: %w", err)
			}
			expired, err := scanExpiredJobs(rows, limit)
			if err != nil {
				return JobRecoveryResult{}, err
			}
			var result JobRecoveryResult
			for _, row := range expired {
				decision := decideExpiredJob(row)
				if err := recoverExpiredJob(
					ctx, transaction, authority, row, decision,
				); err != nil {
					return JobRecoveryResult{}, err
				}
				result.add(decision.state)
			}
			return result, nil
		},
	)
}

func (result *JobRecoveryResult) add(state jobsv1.State) {
	switch state {
	case jobsv1.StateQueued:
		result.requeued++
	case jobsv1.StateSucceeded:
		result.succeeded++
	case jobsv1.StateCancelled:
		result.cancelled++
	case jobsv1.StateDead:
		result.dead++
	}
}
