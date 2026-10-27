package postgres

import (
	"context"
	"fmt"
)

func (repository *workerRepository) reconcileStaleExternalJobs(
	ctx context.Context,
	limit int,
) (finalized int, examined int, resultErr error) {
	jobs, err := repository.listStaleExternalJobs(ctx, limit)
	if err != nil {
		return 0, 0, err
	}
	for _, job := range jobs {
		stored, err := repository.finalizeStaleExternalJob(ctx, job)
		if err != nil {
			return finalized, len(jobs), fmt.Errorf(
				"reconcile stale external job: %w", err,
			)
		}
		if stored {
			finalized++
		}
	}
	return finalized, len(jobs), nil
}
