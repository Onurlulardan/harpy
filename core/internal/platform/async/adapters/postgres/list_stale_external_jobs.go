package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (repository *workerRepository) listStaleExternalJobs(
	ctx context.Context,
	limit int,
) ([]staleExternalJob, error) {
	rows, err := repository.database.Query(ctx, listStaleExternalJobsSQL, limit)
	if err != nil {
		return nil, fmt.Errorf("list stale external jobs: %w", err)
	}
	defer rows.Close()
	jobs := make([]staleExternalJob, 0, limit)
	for rows.Next() {
		var jobID, leaseOwnerID, attemptID uuid.UUID
		var leaseGeneration int64
		var effectStartedAt time.Time
		if err := rows.Scan(
			&jobID, &leaseOwnerID, &leaseGeneration,
			&attemptID, &effectStartedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stale external job: %w", err)
		}
		job, err := newStaleExternalJob(
			repository.identity, jobID, leaseOwnerID, leaseGeneration,
			attemptID, effectStartedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("decode stale external job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stale external jobs: %w", err)
	}
	return jobs, nil
}
