package postgres

import "time"

type JobLease struct {
	owner           *workerIdentity
	claim           JobClaim
	heartbeatAt     time.Time
	until           time.Time
	cancelRequested bool
}

func newJobLease(
	owner *workerIdentity,
	claim JobClaim,
	heartbeatAt time.Time,
	until time.Time,
	cancelRequested bool,
) (JobLease, error) {
	lease := JobLease{
		owner: owner, claim: claim, heartbeatAt: heartbeatAt,
		until: until, cancelRequested: cancelRequested,
	}
	if !claim.valid(owner) || heartbeatAt.IsZero() ||
		!until.Equal(heartbeatAt.Add(jobLeaseDuration)) {
		return JobLease{}, ErrJobClaimInvalid
	}
	return lease, nil
}

func (lease JobLease) HeartbeatAt() time.Time { return lease.heartbeatAt }
func (lease JobLease) Until() time.Time       { return lease.until }
func (lease JobLease) CancelRequested() bool  { return lease.cancelRequested }

func (lease JobLease) valid(owner *workerIdentity) bool {
	return lease.owner == owner && lease.claim.valid(owner) &&
		!lease.heartbeatAt.IsZero() &&
		lease.until.Equal(lease.heartbeatAt.Add(jobLeaseDuration))
}
