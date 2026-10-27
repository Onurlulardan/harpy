package jobs

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type storeFixture struct {
	duplicate, sealed bool
	recurrence        ports.RecurrenceRevision
	future            int
	generated         []ports.GeneratedOccurrence
	reminders         []Reminder
	persist           []bool
	persistIndex      int
	rebuilt, purged   int64
}

func (store *storeFixture) BeginJob(context.Context, Claim) (bool, error) {
	return store.duplicate, nil
}

func (store *storeFixture) SealJob(context.Context, Claim, [32]byte) error {
	store.sealed = true
	return nil
}

func (store *storeFixture) LoadRecurrence(
	context.Context, domain.ID,
) (ports.RecurrenceRevision, error) {
	return store.recurrence, nil
}

func (store *storeFixture) FutureOccurrenceCount(context.Context) (int, error) {
	return store.future, nil
}

func (store *storeFixture) InsertOccurrences(_ context.Context, _ domain.ID,
	values []ports.GeneratedOccurrence,
) (int, error) {
	store.generated = append([]ports.GeneratedOccurrence(nil), values...)
	return len(values), nil
}

func (store *storeFixture) LoadDueReminders(
	context.Context, time.Time, int,
) ([]Reminder, error) {
	return append([]Reminder(nil), store.reminders...), nil
}

func (store *storeFixture) PersistReminder(
	context.Context, Reminder, time.Time,
) (bool, error) {
	inserted := true
	if store.persistIndex < len(store.persist) {
		inserted = store.persist[store.persistIndex]
	}
	store.persistIndex++
	return inserted, nil
}

func (store *storeFixture) RebuildProjection(
	context.Context, domain.ID, time.Time,
) (int64, error) {
	return store.rebuilt, nil
}

func (store *storeFixture) ApplyPurge(
	context.Context, domain.ID, time.Time,
) (int64, error) {
	return store.purged, nil
}

type idFixture struct {
	next domain.ID
	used int
}

func (ids *idFixture) New(time.Time) (domain.ID, error) {
	value := ids.next
	value[15] += byte(ids.used)
	ids.used++
	return value, nil
}

type tzdbFixture struct{}

func (tzdbFixture) Load(context.Context, string, string) (*time.Location, error) {
	return time.UTC, nil
}

func validClaim(kind string) Claim {
	return Claim{JobID: mustJobID("019b2ee0-12e3-7a1a-8d21-5ca5797f2810"),
		TenantID: mustJobID("019b2ee0-12e4-7a1a-8d21-5ca5797f2811"),
		TargetID: mustJobID("019b2ee0-12e5-7a1a-8d21-5ca5797f2812"),
		Kind:     kind, PayloadSHA256: [32]byte{1}, LeaseGeneration: 3,
		Now: time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC)}
}

func mustJobID(value string) domain.ID {
	id, err := domain.ParseID(value)
	if err != nil {
		panic(err)
	}
	return id
}
