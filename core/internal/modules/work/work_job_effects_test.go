package work

import (
	"bytes"
	"testing"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/google/uuid"
)

func TestWorkReminderEffectIsCanonicalAndPolicyPinned(t *testing.T) {
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	claim, err := jobapp.NewPlatformInput(
		uuid.MustParse("019b2ee0-1600-7a1a-8d21-5ca5797f2800"),
		uuid.MustParse("019b2ee0-1601-7a1a-8d21-5ca5797f2801"),
		uuid.MustParse("019b2ee0-1602-7a1a-8d21-5ca5797f2802"),
		asyncmanifestv1.WorkReminderDispatchJobKind, [32]byte{1}, 1, now)
	if err != nil {
		t.Fatal(err)
	}
	reminder, err := jobapp.NewReminderInput(
		uuid.MustParse("019b2ee0-1603-7a1a-8d21-5ca5797f2803"),
		uuid.MustParse("019b2ee0-1604-7a1a-8d21-5ca5797f2804"),
		7, now.Add(-time.Hour), "Review payroll")
	if err != nil {
		t.Fatal(err)
	}
	events, err := workReminderEvents(claim, []jobapp.Reminder{reminder})
	if err != nil || len(events) != 1 {
		t.Fatalf("events = %#v, %v", events, err)
	}
	event := events[0]
	if event.EventType != asyncmanifestv1.WorkReminderDueEventType ||
		event.PayloadSchemaSHA256 != asyncmanifestv1.WorkReminderDueSchemaSHA256 ||
		event.AggregateVersion != 7 || event.EventSequence != 7 ||
		!bytes.Equal(event.Payload, []byte(`{"due_at":"2026-07-19T09:00:00Z","task_id":"019b2ee0-1603-7a1a-8d21-5ca5797f2803","task_version":7,"title":"Review payroll"}`)) {
		t.Fatalf("event = %#v payload=%s", event, event.Payload)
	}
}
