package work

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

type legacyMetadataStoreFixture struct {
	ports.TaskStore
	value domain.LegacyTaskMetadata
	calls int
}

func (store *legacyMetadataStoreFixture) LoadLegacyTaskMetadataForUpdate(
	context.Context, domain.ID,
) (domain.LegacyTaskMetadata, error) {
	store.calls++
	return store.value, nil
}

func TestDecodeWriteTracksPatchPresence(t *testing.T) {
	wire, err := decodeWrite([]byte(
		`{"description":"","due_date":null,"recurrence_rule":null,"content":"note"}`))
	if err != nil || !wire.Present["description"] || !wire.Present["due_date"] ||
		wire.Present["title"] || !wire.Null["due_date"] ||
		!wire.Null["recurrence_rule"] || wire.Null["description"] ||
		wire.Content != "note" {
		t.Fatalf("presence was not preserved: %#v %v", wire.Present, err)
	}
}

func TestLegacyNullPointerFieldsPreserveCurrentValues(t *testing.T) {
	wire, err := decodeWrite([]byte(
		`{"title":null,"form_slug":null,"is_recurring":null,"recurrence_end":null}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"title", "form_slug", "is_recurring", "recurrence_end"} {
		if presentCompatibilityField(wire, field) {
			t.Fatalf("legacy pointer field %s treated JSON null as a patch", field)
		}
	}
	store := &legacyMetadataStoreFixture{}
	replace, _, err := patchLegacyTaskMetadata(context.Background(), store,
		workID(1), true, wire)
	if err != nil || replace || store.calls != 0 {
		t.Fatalf("null recurrence patch mutated metadata: replace=%v calls=%d err=%v",
			replace, store.calls, err)
	}
}

func TestLegacyTaskRelationsDecodeTypedValues(t *testing.T) {
	assigneeID := uuid.New().String()
	checklistID := uuid.New().String()
	body := []byte(`{"assignee_id":"` + assigneeID + `","assignee_type":"USER",` +
		`"labels":["Urgent"],"checklist":[{"id":"` + checklistID +
		`","text":"Review","checked":true},{"id":"cl-1","text":"Send","checked":false}]}`)
	wire, err := decodeWrite(body)
	if err != nil {
		t.Fatal(err)
	}
	value, err := taskRelations(wire, true, false)
	if err != nil || !value.ReplaceAssignees || !value.ReplaceLabels ||
		!value.ReplaceChecklist || len(value.Assignees) != 1 || len(value.Labels) != 1 ||
		len(value.Checklist) != 2 || value.Checklist[0].ID == nil || value.Checklist[1].ID != nil {
		t.Fatalf("unexpected relation decode: %#v %v", value, err)
	}
}

func TestLegacyTaskTypeAcceptsCanonicalFormKindsOnly(t *testing.T) {
	kind, err := taskKind(writeWire{TaskType: "FORM_FILL"}, true)
	if err != nil || kind != ports.TaskFormFill {
		t.Fatalf("FORM_FILL should retain its canonical kind, got %q %v", kind, err)
	}
	_, err = taskKind(writeWire{TaskType: "APPROVAL"}, true)
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatalf("APPROVAL should be rejected, got %v", err)
	}
}

func TestLegacyTaskMetadataPreservesRecurrenceAndIgnoresUnsupportedFilterWrite(t *testing.T) {
	recurring := true
	wire := writeWire{IsRecurring: &recurring, RecurrenceRule: "FREQ=WEEKLY",
		RecurrenceEnd: "2026-12-31T00:00:00Z",
		FormFilter:    []byte(`{"status":"open"}`)}
	value, err := createLegacyTaskMetadata(wire, true)
	if err != nil || value == nil || !value.IsRecurring ||
		value.RecurrenceRule != "FREQ=WEEKLY" || value.RecurrenceEnd == nil ||
		value.FormFilter != nil {
		t.Fatalf("legacy metadata was not preserved: value=%+v err=%v", value, err)
	}
	if canonical, err := createLegacyTaskMetadata(wire, false); err != nil || canonical != nil {
		t.Fatalf("canonical task leaked compatibility metadata: value=%+v err=%v", canonical, err)
	}
}

func TestLegacyTaskMetadataPatchPreservesAbsentAndClearsPresent(t *testing.T) {
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	store := &legacyMetadataStoreFixture{value: domain.LegacyTaskMetadata{
		IsRecurring: true, RecurrenceRule: "FREQ=WEEKLY", RecurrenceEnd: &end,
		FormFilter: []byte(`{"status":"open"}`)}}
	replace, _, err := patchLegacyTaskMetadata(context.Background(), store,
		workID(1), true, writeWire{Present: map[string]bool{}})
	if err != nil || replace || store.calls != 0 {
		t.Fatalf("absent metadata patch did not preserve: replace=%v calls=%d err=%v",
			replace, store.calls, err)
	}
	active := false
	replace, value, err := patchLegacyTaskMetadata(context.Background(), store,
		workID(1), true, writeWire{Present: map[string]bool{
			"is_recurring": true, "recurrence_rule": true,
			"recurrence_end": true, "form_filter": true}, IsRecurring: &active,
			FormFilter: []byte(`{}`)})
	if err != nil || !replace || store.calls != 1 || value.IsRecurring ||
		value.RecurrenceRule != "" || value.RecurrenceEnd != nil ||
		string(value.FormFilter) != `{"status":"open"}` {
		t.Fatalf("explicit metadata clear drifted: value=%+v calls=%d err=%v",
			value, store.calls, err)
	}
}

func TestLegacyTaskMetadataIgnoresInvalidRecurrenceEndLikeV1(t *testing.T) {
	created, err := createLegacyTaskMetadata(writeWire{
		RecurrenceEnd: "not-an-rfc3339-time", FormFilter: []byte(`{"ignored":true}`),
	}, true)
	if err != nil || created == nil || created.RecurrenceEnd != nil || created.FormFilter != nil {
		t.Fatalf("legacy create invalid end was not ignored: value=%+v err=%v", created, err)
	}
	current := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	store := &legacyMetadataStoreFixture{value: domain.LegacyTaskMetadata{
		RecurrenceEnd: &current,
	}}
	wire, err := decodeWrite([]byte(
		`{"recurrence_end":"not-an-rfc3339-time","form_filter":null}`))
	if err != nil {
		t.Fatal(err)
	}
	replace, patched, err := patchLegacyTaskMetadata(context.Background(), store,
		workID(1), true, wire)
	if err != nil || !replace || patched.RecurrenceEnd == nil ||
		!patched.RecurrenceEnd.Equal(current) || patched.FormFilter != nil {
		t.Fatalf("legacy update invalid end did not preserve current: value=%+v err=%v",
			patched, err)
	}
}

func TestTaskCommentBodyPreservesCompatibilityContent(t *testing.T) {
	wire := writeWire{Body: "canonical", Content: "legacy"}
	if got := taskCommentBody(wire, true); got != "legacy" {
		t.Fatalf("compatibility content drifted: %q", got)
	}
	if got := taskCommentBody(wire, false); got != "canonical" {
		t.Fatalf("canonical body drifted: %q", got)
	}
}

func TestPatchedTaskTimeDistinguishesAbsentAndClear(t *testing.T) {
	now := time.Now().UTC()
	value, err := patchedTaskTime(writeWire{Present: map[string]bool{}}, true,
		"due_at", "due_date", &now)
	if err != nil || value == nil || !value.Equal(now) {
		t.Fatal("absent due date did not preserve current value")
	}
	value, err = patchedTaskTime(writeWire{Present: map[string]bool{"due_date": true}},
		true, "due_at", "due_date", &now)
	if err != nil || value != nil {
		t.Fatal("explicit empty due date did not clear value")
	}
}
