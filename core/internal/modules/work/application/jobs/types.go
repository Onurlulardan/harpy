package jobs

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type Claim struct {
	JobID, TenantID, TargetID domain.ID
	Kind                      string
	PayloadSHA256             [32]byte
	LeaseGeneration           int64
	Now                       time.Time
}

func (claim Claim) Valid() bool {
	return claim.JobID.UUIDv7() && claim.TenantID.Valid() && claim.TargetID.Valid() &&
		claim.Kind != "" && claim.PayloadSHA256 != [32]byte{} &&
		claim.LeaseGeneration > 0 && !claim.Now.IsZero() && claim.Now.Location() == time.UTC
}

type Reminder struct {
	TaskID      domain.ID
	TaskVersion int64
	DueAt       time.Time
	Title       string
	EventID     domain.ID
}

type Result struct {
	Duplicate   bool
	Occurrences int
	Reminders   []Reminder
	Repaired    int64
	Purged      int64
}

type Store interface {
	BeginJob(context.Context, Claim) (bool, error)
	SealJob(context.Context, Claim, [32]byte) error
	LoadRecurrence(context.Context, domain.ID) (ports.RecurrenceRevision, error)
	FutureOccurrenceCount(context.Context) (int, error)
	InsertOccurrences(context.Context, domain.ID, []ports.GeneratedOccurrence) (int, error)
	LoadDueReminders(context.Context, time.Time, int) ([]Reminder, error)
	PersistReminder(context.Context, Reminder, time.Time) (bool, error)
	RebuildProjection(context.Context, domain.ID, time.Time) (int64, error)
	ApplyPurge(context.Context, domain.ID, time.Time) (int64, error)
}
