package humantask

import (
	"errors"
	"time"
)

var ErrInvalidVerification = errors.New("IAM human task verification is invalid")

func (request VerificationRequest) Validate() error {
	if !stable(request.TenantID) || !stable(request.ActorID) ||
		!request.Candidate.valid() || !validAction(request.Action) ||
		!canonicalTime(request.VerifyAt) {
		return ErrInvalidVerification
	}
	return nil
}

func (evidence VerificationEvidence) Validate() error {
	request := VerificationRequest{
		TenantID: evidence.TenantID, ActorID: evidence.ActorID,
		Candidate: evidence.Candidate, Action: evidence.Action,
		VerifyAt: evidence.VerifiedAt,
	}
	if request.Validate() != nil || evidence.AuthorizationVersion < 1 ||
		!stable(evidence.PermissionID) || evidence.EvidenceSHA256 == [32]byte{} {
		return ErrInvalidVerification
	}
	return nil
}

func (evidence VerificationEvidence) Matches(request VerificationRequest) bool {
	return evidence.Validate() == nil && request.Validate() == nil &&
		evidence.TenantID == request.TenantID && evidence.ActorID == request.ActorID &&
		evidence.Candidate == request.Candidate && evidence.Action == request.Action &&
		evidence.VerifiedAt.Equal(request.VerifyAt)
}

func (request AssignmentRequest) Validate() error {
	if !stable(request.TenantID) || !request.Assignment.valid() ||
		!canonicalTime(request.VerifyAt) {
		return ErrInvalidVerification
	}
	return nil
}

func (evidence AssignmentEvidence) Validate() error {
	request := AssignmentRequest{TenantID: evidence.TenantID,
		Assignment: evidence.Assignment, VerifyAt: evidence.VerifiedAt}
	if request.Validate() != nil || evidence.Revision < 1 ||
		evidence.EvidenceSHA256 == [32]byte{} {
		return ErrInvalidVerification
	}
	return nil
}

func (evidence AssignmentEvidence) Matches(request AssignmentRequest) bool {
	return evidence.Validate() == nil && request.Validate() == nil &&
		evidence.TenantID == request.TenantID &&
		evidence.Assignment == request.Assignment &&
		evidence.VerifiedAt.Equal(request.VerifyAt)
}

func (candidate Candidate) valid() bool {
	return stable(candidate.ID) && (candidate.Kind == "user" ||
		candidate.Kind == "role" || candidate.Kind == "organization")
}

func validAction(value string) bool {
	switch value {
	case "claim", "decide", "reassign", "expire", "cancel":
		return true
	default:
		return false
	}
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0 &&
		value.Year() >= 1 && value.Year() <= 9999
}

func stable(value [16]byte) bool {
	return value != [16]byte{} && value[8]&0xc0 == 0x80
}
