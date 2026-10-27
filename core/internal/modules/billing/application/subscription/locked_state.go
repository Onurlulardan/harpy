package subscriptionapp

import "time"

type LockedState struct {
	Subscription  SubscriptionSnapshot
	Receipt       *OperationReceipt
	TargetPrice   *TargetPrice
	BoundaryJobs  []JobIntent
	PendingChange *SubscriptionChange
	TrialClaim    *TrialClaim
	Boundary      *BoundarySnapshot
}

func (locked LockedState) valid() bool {
	return locked.Subscription.valid() &&
		(locked.Receipt == nil || locked.Receipt.valid()) &&
		(locked.TargetPrice == nil || locked.TargetPrice.valid()) &&
		locked.validBoundaryJobs() &&
		(locked.PendingChange == nil || locked.PendingChange.valid()) &&
		(locked.TrialClaim == nil || locked.TrialClaim.valid()) &&
		(locked.Boundary == nil || locked.Boundary.valid())
}

func (locked LockedState) validBoundaryJobs() bool {
	seen := map[BoundaryKind]bool{}
	for _, job := range locked.BoundaryJobs {
		if !job.valid() || job.SubscriptionID != locked.Subscription.ID || seen[job.Boundary] {
			return false
		}
		seen[job.Boundary] = true
	}
	return true
}

func (locked LockedState) boundaryJob(kind BoundaryKind) (JobIntent, bool) {
	for _, job := range locked.BoundaryJobs {
		if job.Boundary == kind {
			return job, true
		}
	}
	return JobIntent{}, false
}

func boundaryJobCancellations(locked LockedState, now time.Time, reason string) []JobCancellation {
	result := make([]JobCancellation, 0, len(locked.BoundaryJobs))
	for _, job := range locked.BoundaryJobs {
		result = append(result, JobCancellation{ID: job.ID, CancelledAt: now, Reason: reason})
	}
	return result
}
