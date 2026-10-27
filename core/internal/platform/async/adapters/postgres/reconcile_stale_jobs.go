package postgres

import (
	"context"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func (repository *workerRepository) ReconcileStaleJobs(
	ctx context.Context,
	limit int,
) (int, error) {
	if ctx == nil {
		return 0, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return 0, ErrWorkerStoreUnavailable
	}
	if limit < 1 || limit > jobsv1.MaximumClaimBatch {
		return 0, ErrStaleJobLimit
	}
	finalized, examined, err := repository.reconcileStaleExternalJobs(ctx, limit)
	if err != nil {
		return finalized, err
	}
	remaining := limit - examined
	if remaining == 0 {
		return finalized, nil
	}
	jobs, err := repository.listStaleJobs(ctx, remaining)
	if err != nil {
		return finalized, err
	}
	parked := finalized
	for _, job := range jobs {
		stored, err := repository.parkStaleJob(ctx, job)
		if err != nil {
			return parked, fmt.Errorf("reconcile stale job: %w", err)
		}
		if stored {
			parked++
		}
	}
	return parked, nil
}
