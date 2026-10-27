package application

import (
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type RetentionJobClaim struct {
	JobID            uuid.UUID
	AttemptID        uuid.UUID
	Authority        RetentionAuthority
	Kind             string
	PayloadRevision  int
	PayloadReference string
	PayloadSHA256    string
	LeaseGeneration  int64
	StartedAt        time.Time
}

func (claim RetentionJobClaim) cycle() (uuid.UUID, time.Time, error) {
	environmentID, cycleAt, err := ParseRetentionReference(claim.PayloadReference)
	if err != nil || !validV7(claim.JobID) || !validV7(claim.AttemptID) ||
		claim.Kind != RetentionJobKind || claim.PayloadRevision != RetentionPayloadRevision ||
		claim.LeaseGeneration < 1 || !validRetentionTime(claim.StartedAt) ||
		cycleAt.After(claim.StartedAt) || !claim.Authority.validFor(environmentID) {
		return uuid.Nil, time.Time{}, domain.ErrForbidden
	}
	_, expectedDigest, err := RetentionReference(environmentID, cycleAt)
	if err != nil || claim.PayloadSHA256 != expectedDigest {
		return uuid.Nil, time.Time{}, domain.ErrInvalid
	}
	return environmentID, cycleAt, nil
}

type RetentionPrune struct {
	ReceiptID     uuid.UUID
	EnvironmentID uuid.UUID
	DeleteBefore  time.Time
	BatchLimit    int
	ObservedAt    time.Time
	Authority     RetentionAuthority
}

func (request RetentionPrune) Validate() error {
	if !validV7(request.ReceiptID) || request.EnvironmentID == uuid.Nil ||
		!validRetentionTime(request.DeleteBefore) ||
		!validRetentionTime(request.ObservedAt) ||
		request.DeleteBefore != request.ObservedAt.Add(-RetentionAge) ||
		request.BatchLimit != RetentionBatchLimit ||
		!request.Authority.validFor(request.EnvironmentID) {
		return domain.ErrForbidden
	}
	return nil
}

type RetentionSuccessor struct {
	JobID            uuid.UUID
	RunAt            time.Time
	PayloadReference string
	PayloadSHA256    string
	DedupeKey        string
	Priority         int32
}

type RetentionOutcome struct {
	DeletedCount int
	HasBacklog   bool
	Successor    RetentionSuccessor
}

func (outcome RetentionOutcome) Validate(environmentID uuid.UUID) error {
	next := outcome.Successor
	nextEnvironment, nextCycle, err := ParseRetentionReference(next.PayloadReference)
	if err != nil || nextEnvironment != environmentID ||
		outcome.DeletedCount < 0 || outcome.DeletedCount > RetentionBatchLimit ||
		outcome.HasBacklog != (outcome.DeletedCount == RetentionBatchLimit) ||
		!validV7(next.JobID) || !validRetentionTime(next.RunAt) ||
		strings.TrimSpace(next.DedupeKey) != next.DedupeKey ||
		next.DedupeKey != "usage:retention:"+next.JobID.String() || next.Priority != 0 {
		return domain.ErrUnavailable
	}
	_, digest, err := RetentionReference(environmentID, nextCycle)
	if err != nil || next.PayloadSHA256 != digest {
		return domain.ErrUnavailable
	}
	return nil
}

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7) &&
		value.Variant() == uuid.RFC4122
}
