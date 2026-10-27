package api

import "time"

type ClaimAttemptRequest struct {
	OperationID      UUID
	AttemptID        UUID
	LeaseOwner       string
	LeaseTokenSHA256 string
	ExpectedRevision int64
	LeaseDuration    time.Duration
}

type AttemptLease struct {
	operationID UUID
	attemptID   UUID
	attempt     int64
	revision    int64
	leaseUntil  time.Time
}

func NewAttemptLease(
	operationID UUID,
	attemptID UUID,
	attempt int64,
	revision int64,
	leaseUntil time.Time,
) (AttemptLease, error) {
	if operationID == NilUUID || attemptID == NilUUID || attempt < 1 || revision < 1 ||
		!microsecondUTC(leaseUntil) {
		return AttemptLease{}, ErrOnboardingInvalid
	}
	return AttemptLease{operationID: operationID, attemptID: attemptID,
		attempt: attempt, revision: revision, leaseUntil: leaseUntil}, nil
}

func (lease AttemptLease) OperationID() UUID     { return lease.operationID }
func (lease AttemptLease) AttemptID() UUID       { return lease.attemptID }
func (lease AttemptLease) Attempt() int64        { return lease.attempt }
func (lease AttemptLease) Revision() int64       { return lease.revision }
func (lease AttemptLease) LeaseUntil() time.Time { return lease.leaseUntil }
