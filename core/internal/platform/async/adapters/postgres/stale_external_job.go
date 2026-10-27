package postgres

import (
	"time"

	"github.com/google/uuid"
)

type staleExternalJob struct {
	owner           *workerIdentity
	jobID           uuid.UUID
	leaseOwnerID    uuid.UUID
	leaseGeneration int64
	attemptID       uuid.UUID
	effectStartedAt time.Time
}

func newStaleExternalJob(
	owner *workerIdentity,
	jobID uuid.UUID,
	leaseOwnerID uuid.UUID,
	leaseGeneration int64,
	attemptID uuid.UUID,
	effectStartedAt time.Time,
) (staleExternalJob, error) {
	job := staleExternalJob{
		owner: owner, jobID: jobID, leaseOwnerID: leaseOwnerID,
		leaseGeneration: leaseGeneration, attemptID: attemptID,
		effectStartedAt: effectStartedAt,
	}
	if !job.valid(owner) {
		return staleExternalJob{}, ErrStaleJobInvalid
	}
	return job, nil
}

func (job staleExternalJob) valid(owner *workerIdentity) bool {
	return owner != nil && owner.marker == 1 && job.owner == owner &&
		job.jobID != uuid.Nil && job.leaseOwnerID != uuid.Nil &&
		job.leaseGeneration >= 1 && job.attemptID != uuid.Nil &&
		!job.effectStartedAt.IsZero()
}

func (job staleExternalJob) arguments() []any {
	return []any{
		job.jobID, job.leaseOwnerID, job.leaseGeneration,
		job.attemptID, job.effectStartedAt,
	}
}
