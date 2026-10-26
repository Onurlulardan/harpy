package retention

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Counts struct {
	ContentRedacted       int64
	AttemptsPurged        int64
	ReceiptsPurged        int64
	EffectsPurged         int64
	DeliveriesPurged      int64
	MessagesPurged        int64
	HintsPurged           int64
	PlatformRecordsPurged int64
}

func (counts Counts) valid() bool {
	return counts.ContentRedacted >= 0 && counts.AttemptsPurged >= 0 &&
		counts.ReceiptsPurged >= 0 && counts.EffectsPurged >= 0 &&
		counts.DeliveriesPurged >= 0 && counts.MessagesPurged >= 0 &&
		counts.HintsPurged >= 0 && counts.PlatformRecordsPurged >= 0
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
	HasRemaining       bool
	CheckpointRevision int64
}

func (outcome Outcome) Validate(tenantID uuid.UUID) error {
	reference, digest, err := Reference(tenantID)
	successor := outcome.Successor
	if err != nil || !outcome.Counts.valid() || outcome.CheckpointRevision < 1 ||
		successor.JobID == uuid.Nil || successor.JobID.Version() != uuid.Version(7) ||
		successor.RunAt.IsZero() || successor.RunAt.Location() != time.UTC ||
		successor.RunAt.Nanosecond()%int(time.Microsecond) != 0 ||
		successor.PayloadReference != reference || successor.PayloadSHA256 != digest ||
		strings.TrimSpace(successor.DedupeKey) != successor.DedupeKey ||
		len(successor.DedupeKey) < 1 || len(successor.DedupeKey) > 512 {
		return ErrOutcomeInvalid
	}
	return nil
}
