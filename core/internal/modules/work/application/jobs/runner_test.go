package jobs

import (
	"context"
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestRunnerMaterializesPinnedRecurrenceAndSealsReceipt(t *testing.T) {
	store := &storeFixture{recurrence: ports.RecurrenceRevision{Revision: 2,
		Contract: domain.RecurrenceContract, DTSTART: "20260718T090000",
		TimeZone: "UTC", TZDBVersion: "2026a", RRULE: "FREQ=DAILY;COUNT=4"}}
	ids := &idFixture{next: mustJobID("019b2ee0-1300-7a1a-8d21-5ca5797f2800")}
	runner, err := NewRunner(store, tzdbFixture{}, ids)
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.RunJob(context.Background(),
		validClaim(asyncmanifestv1.WorkRecurrenceMaterializeJobKind))
	if err != nil || result.Occurrences != 4 || len(store.generated) != 4 ||
		!store.sealed {
		t.Fatalf("materialize = %#v generated=%d sealed=%v err=%v",
			result, len(store.generated), store.sealed, err)
	}
	for _, generated := range store.generated {
		if !generated.ID.UUIDv7() || generated.Occurrence.TZDBVersion != "2026a" {
			t.Fatalf("invalid generated occurrence: %#v", generated)
		}
	}
}

func TestRunnerReminderReceiptsSuppressDuplicateEffects(t *testing.T) {
	store := &storeFixture{reminders: []Reminder{
		{TaskID: mustJobID("019b2ee0-1310-7a1a-8d21-5ca5797f2810"),
			TaskVersion: 2, DueAt: validClaim("x").Now, Title: "First"},
		{TaskID: mustJobID("019b2ee0-1311-7a1a-8d21-5ca5797f2811"),
			TaskVersion: 4, DueAt: validClaim("x").Now, Title: "Second"},
	}, persist: []bool{true, false}}
	ids := &idFixture{next: mustJobID("019b2ee0-1320-7a1a-8d21-5ca5797f2820")}
	runner, err := NewRunner(store, tzdbFixture{}, ids)
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.RunJob(context.Background(),
		validClaim(asyncmanifestv1.WorkReminderDispatchJobKind))
	if err != nil || len(result.Reminders) != 1 || !result.Reminders[0].EventID.UUIDv7() ||
		store.persistIndex != 2 || !store.sealed {
		t.Fatalf("reminder result = %#v persisted=%d sealed=%v err=%v",
			result, store.persistIndex, store.sealed, err)
	}
}

func TestRunnerDispatchesRepairPurgeAndRejectsUnknown(t *testing.T) {
	tests := []struct {
		kind   string
		value  int64
		result func(Result) int64
	}{
		{asyncmanifestv1.WorkProjectionRepairJobKind, 7, func(value Result) int64 {
			return value.Repaired
		}},
		{asyncmanifestv1.WorkPurgeExecuteJobKind, 1, func(value Result) int64 {
			return value.Purged
		}},
	}
	for _, test := range tests {
		store := &storeFixture{rebuilt: test.value, purged: test.value}
		runner, _ := NewRunner(store, tzdbFixture{}, &idFixture{})
		result, err := runner.RunJob(context.Background(), validClaim(test.kind))
		if err != nil || test.result(result) != test.value || !store.sealed {
			t.Fatalf("RunJob(%s) = %#v, %v", test.kind, result, err)
		}
	}
	store := &storeFixture{}
	runner, _ := NewRunner(store, tzdbFixture{}, &idFixture{})
	if _, err := runner.RunJob(context.Background(), validClaim("work.unknown")); err == nil || store.sealed {
		t.Fatal("unknown job was accepted")
	}
}

func TestRunnerRecognizesCommittedReplay(t *testing.T) {
	store := &storeFixture{duplicate: true}
	runner, _ := NewRunner(store, tzdbFixture{}, &idFixture{})
	result, err := runner.RunJob(context.Background(),
		validClaim(asyncmanifestv1.WorkReminderDispatchJobKind))
	if err != nil || !result.Duplicate || store.sealed {
		t.Fatalf("duplicate = %#v sealed=%v err=%v", result, store.sealed, err)
	}
}
