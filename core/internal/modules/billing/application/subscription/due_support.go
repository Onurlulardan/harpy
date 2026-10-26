package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

func boundaryCancellationsExcept(
	locked LockedState,
	jobID uuid.UUID,
	now time.Time,
	reason string,
) []JobCancellation {
	result := make([]JobCancellation, 0, len(locked.BoundaryJobs))
	for _, job := range locked.BoundaryJobs {
		if job.ID != jobID {
			result = append(result, JobCancellation{ID: job.ID, CancelledAt: now, Reason: reason})
		}
	}
	return result
}

func dueDigest(request ExecuteDueBoundaryRequest) [32]byte {
	writer := newDigestWriter()
	writer.text(OperationExecuteDue)
	writer.identifier(request.SubscriptionID)
	writer.identifier(request.JobID)
	writer.identifier(request.EffectID)
	writer.integer(request.ExpectedRevision)
	writer.integer(request.LeaseGeneration)
	writer.text(string(request.Boundary))
	writer.instant(request.DueAt)
	return writer.sum()
}
