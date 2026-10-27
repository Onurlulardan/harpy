package v1adoption

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestRecurringTaskProjectionUsesValidCalendarDefaults(t *testing.T) {
	due := time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC)
	end := due.Add(30 * 24 * time.Hour)
	task := LegacyTask{ID: fixtureID(10), BoardID: fixtureID(11),
		CreatedBy: fixtureID(12), Title: "Recurring task", IsRecurring: true,
		RecurrenceRule: "FREQ=WEEKLY", DueDate: &due, RecurrenceEnd: &end,
		CreatedAt: due.Add(-24 * time.Hour), UpdatedAt: due.Add(-time.Hour)}
	source := fixtureRecord(KindTask, task.ID, task)
	service := fixtureService(nil, &fixtureTarget{})
	plan := Plan{}
	if err := service.transformTaskRecurrence(context.Background(), fixtureTenantScope(),
		&plan, source, LegacyBoard{}, task); err != nil {
		t.Fatal(err)
	}
	if len(plan.Issues) != 0 || len(plan.Records) != 2 {
		t.Fatalf("recurrence projection records=%d issues=%+v", len(plan.Records), plan.Issues)
	}
	for _, record := range plan.Records {
		if record.Kind != TargetCalendar {
			continue
		}
		var document TargetCalendarDocument
		if json.Unmarshal(record.Document, &document) != nil || document.Color != "#64748b" ||
			document.Audience != "BOARD" || document.BoardID != task.BoardID.String() {
			t.Fatalf("calendar projection defaults drifted: %+v", document)
		}
		return
	}
	t.Fatal("recurrence projection omitted its calendar event")
}
