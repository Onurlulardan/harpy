package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *workerRepository) listStaleJobs(
	ctx context.Context,
	limit int,
) ([]staleJob, error) {
	rows, err := repository.database.Query(ctx, listStaleJobsSQL, limit)
	if err != nil {
		return nil, fmt.Errorf("list stale jobs: %w", err)
	}
	defer rows.Close()
	jobs := make([]staleJob, 0, limit)
	for rows.Next() {
		var jobID uuid.UUID
		var state, kind, payloadReference, payloadSHA256 string
		var leaseOwner pgtype.UUID
		var leaseGeneration int64
		var payloadRevision int32
		if err := rows.Scan(
			&jobID, &state, &leaseOwner, &leaseGeneration, &kind,
			&payloadRevision, &payloadReference, &payloadSHA256,
		); err != nil {
			return nil, fmt.Errorf("scan stale job: %w", err)
		}
		ownerID := uuid.Nil
		if leaseOwner.Valid {
			ownerID = uuid.UUID(leaseOwner.Bytes)
		}
		job, err := newStaleJob(
			repository.identity, jobID, state, ownerID, leaseOwner.Valid,
			leaseGeneration, kind, int(payloadRevision), payloadReference, payloadSHA256,
		)
		if err != nil {
			return nil, fmt.Errorf("decode stale job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stale jobs: %w", err)
	}
	return jobs, nil
}
