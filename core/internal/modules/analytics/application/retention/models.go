package retention

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Counts struct {
	Drafts       int64
	Versions     int64
	Releases     int64
	Executions   int64
	Audits       int64
	Capabilities int64
}

type Claim struct {
	JobID            uuid.UUID
	AttemptID        uuid.UUID
	TenantID         uuid.UUID
	Kind             string
	PayloadRevision  int
	PayloadReference string
	PayloadSHA256    string
}

func (claim Claim) validate() error {
	if !validV7(claim.JobID) || !validV7(claim.AttemptID) ||
		claim.TenantID == uuid.Nil || claim.Kind != JobKind ||
		claim.PayloadRevision != PayloadRevision {
		return ErrClaimInvalid
	}
	reference, digest, err := Reference(claim.TenantID)
	if err != nil || claim.PayloadReference != reference || claim.PayloadSHA256 != digest {
		return ErrClaimInvalid
	}
	return nil
}

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7) &&
		value.Variant() == uuid.RFC4122
}

func (counts Counts) valid() bool {
	values := []int64{counts.Drafts, counts.Versions, counts.Releases,
		counts.Executions, counts.Audits, counts.Capabilities}
	var total int64
	for _, value := range values {
		if value < 0 {
			return false
		}
		total += value
	}
	return total <= BatchLimit
}

type Successor struct {
	JobID            uuid.UUID
	RunAt            time.Time
	PayloadReference string
	PayloadSHA256    string
	DedupeKey        string
	Priority         int32
}

type Outcome struct {
	Counts             Counts
	Successor          Successor
	LegalHoldBlocked   bool
	HasRemaining       bool
	CheckpointRevision int64
}

func (outcome Outcome) Validate(tenantID uuid.UUID) error {
	reference, digest, err := Reference(tenantID)
	next := outcome.Successor
	if err != nil || !outcome.Counts.valid() || outcome.CheckpointRevision < 1 ||
		(outcome.LegalHoldBlocked && (outcome.Counts != Counts{} || outcome.HasRemaining)) ||
		!validV7(next.JobID) || next.RunAt.IsZero() || next.RunAt.Location() != time.UTC ||
		next.RunAt.Nanosecond()%int(time.Microsecond) != 0 ||
		next.PayloadReference != reference || next.PayloadSHA256 != digest ||
		strings.TrimSpace(next.DedupeKey) != next.DedupeKey ||
		len(next.DedupeKey) < 1 || len(next.DedupeKey) > 256 {
		return ErrOutcomeInvalid
	}
	return nil
}
