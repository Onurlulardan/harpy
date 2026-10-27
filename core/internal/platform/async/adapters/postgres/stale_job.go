package postgres

import (
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

type staleJob struct {
	owner            *workerIdentity
	jobID            uuid.UUID
	state            string
	leaseOwnerID     uuid.UUID
	hasLeaseOwner    bool
	leaseGeneration  int64
	kind             string
	payloadRevision  int
	payloadReference string
	payloadSHA256    string
}

func newStaleJob(
	owner *workerIdentity,
	jobID uuid.UUID,
	state string,
	leaseOwnerID uuid.UUID,
	hasLeaseOwner bool,
	leaseGeneration int64,
	kind string,
	payloadRevision int,
	payloadReference string,
	payloadSHA256 string,
) (staleJob, error) {
	job := staleJob{
		owner: owner, jobID: jobID, state: state,
		leaseOwnerID: leaseOwnerID, hasLeaseOwner: hasLeaseOwner,
		leaseGeneration: leaseGeneration, kind: kind,
		payloadRevision: payloadRevision, payloadReference: payloadReference,
		payloadSHA256: payloadSHA256,
	}
	if !job.valid(owner) {
		return staleJob{}, ErrStaleJobInvalid
	}
	return job, nil
}

func staleJobFromClaim(owner *workerIdentity, claim JobClaim) (staleJob, error) {
	if !claim.valid(owner) {
		return staleJob{}, ErrJobClaimInvalid
	}
	return newStaleJob(
		owner, claim.jobID, string(jobsv1.StateRunning), claim.leaseOwnerID, true,
		claim.leaseGeneration, claim.kind, claim.payloadRevision,
		claim.payloadReference, claim.payloadSHA256,
	)
}

func (job staleJob) valid(owner *workerIdentity) bool {
	leaseValid := job.state == string(jobsv1.StateQueued) &&
		!job.hasLeaseOwner && job.leaseOwnerID == uuid.Nil && job.leaseGeneration >= 0 ||
		job.state == string(jobsv1.StateRunning) && job.hasLeaseOwner &&
			job.leaseOwnerID != uuid.Nil && job.leaseGeneration >= 1
	return owner != nil && owner.marker == 1 && job.owner == owner &&
		job.jobID != uuid.Nil && leaseValid && claimKindPattern.MatchString(job.kind) &&
		job.payloadRevision >= 1 && len(job.payloadReference) >= 1 &&
		len(job.payloadReference) <= jobsv1.MaximumReferenceBytes &&
		claimDigestPattern.MatchString(job.payloadSHA256)
}

func (job staleJob) leaseOwnerArgument() any {
	if !job.hasLeaseOwner {
		return nil
	}
	return job.leaseOwnerID
}
