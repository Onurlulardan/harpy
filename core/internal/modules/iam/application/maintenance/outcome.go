package maintenance

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

func (value Successor) Validate() error {
	if !stableID(value.JobID) || value.JobID.Version() != uuid.Version(7) ||
		value.RunAt.IsZero() || value.RunAt.Location() != time.UTC ||
		!boundedText(value.PayloadReference, 1024) ||
		digestReference(value.PayloadReference) != value.PayloadSHA256 ||
		!boundedText(value.DedupeKey, 256) {
		return ErrOutcomeInvalid
	}
	return nil
}

type Outcome struct {
	Successor *Successor
}

func (outcome Outcome) Validate(kind Kind) error {
	if !kind.Valid() || kind.Recurring() && outcome.Successor == nil {
		return ErrOutcomeInvalid
	}
	if outcome.Successor != nil {
		return outcome.Successor.Validate()
	}
	return nil
}

func boundedText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}
