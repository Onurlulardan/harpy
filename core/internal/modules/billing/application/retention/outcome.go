package retention

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Successor struct {
	JobID            uuid.UUID
	RunAt            time.Time
	PayloadReference string
	PayloadSHA256    string
	DedupeKey        string
	Priority         int32
}

func (successor Successor) validate(tenantID uuid.UUID) error {
	reference, digest, err := Reference(tenantID)
	if err != nil || successor.JobID == uuid.Nil ||
		successor.JobID.Version() != uuid.Version(7) || successor.RunAt.IsZero() ||
		successor.RunAt.Location() != time.UTC ||
		successor.RunAt.Nanosecond()%int(time.Microsecond) != 0 ||
		successor.PayloadReference != reference || successor.PayloadSHA256 != digest ||
		!bounded(successor.DedupeKey, 256) {
		return ErrOutcomeInvalid
	}
	return nil
}

type Outcome struct {
	Successor          Successor
	RawBodiesRedacted  int64
	RawRecordsPurged   int64
	EntitlementPurged  int64
	FinancialPurged    int64
	LegalHoldBlocked   int64
	CheckpointRevision int64
}

func (outcome Outcome) Validate(tenantID uuid.UUID) error {
	if outcome.RawBodiesRedacted < 0 || outcome.RawRecordsPurged < 0 ||
		outcome.EntitlementPurged < 0 || outcome.FinancialPurged < 0 ||
		outcome.LegalHoldBlocked < 0 || outcome.CheckpointRevision < 1 {
		return ErrOutcomeInvalid
	}
	return outcome.Successor.validate(tenantID)
}

func bounded(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}
