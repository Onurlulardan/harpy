//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestWorkPG18SQLCutoverPreservesTaskCompatibilityMetadata(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startWorkPG18(t, ctx)
	prepareWorkSchemas(t, ctx, database.owner)
	authority := prepareProtectedWorkAuthority(t, ctx, database.owner, "work_adopted_a")
	assertProtectedTaskMetadataDocumentGuard(t, ctx, database.owner, authority.schema)

	batches, expected := newProtectedTaskMetadataBatches(authority), []int64{1, 1, 0, 1, 0, 0}
	for index, batch := range batches {
		applied, _, _, err := executeProtectedWorkBatch(ctx, database.owner,
			authority, authority.schema, batch, true)
		if err != nil || applied != expected[index] {
			t.Fatalf("protected task metadata batch %s = %d want=%d err=%v",
				batch.checkpointKey, applied, expected[index], err)
		}
	}
	assertProtectedTaskMetadata(t, ctx, database.owner, authority.schema)
}

func newProtectedTaskMetadataBatches(a protectedWorkAuthority) []protectedWorkBatch {
	boardID := "41000000-0000-4000-8000-000000000302"
	columnID := "41000000-0000-4000-8000-000000000304"
	taskID := "41000000-0000-4000-8000-000000000306"
	creatorID := "41000000-0000-4000-8000-000000000303"
	records := []struct {
		sourceKind, sourceID, targetKind, targetID string
		record                                     map[string]any
	}{
		{"task_boards", "41000000-0000-4000-8000-000000000301", "BOARD", boardID,
			map[string]any{"ArchivedAt": nil, "CreatedAt": "2026-07-21T09:00:00Z",
				"CreatorID": creatorID, "Description": "Task metadata board",
				"Lifecycle": "ACTIVE", "Name": "Task metadata",
				"UpdatedAt": "2026-07-21T09:00:00Z"}},
		{"task_board_columns", "41000000-0000-4000-8000-000000000305", "COLUMN", columnID,
			map[string]any{"BoardID": boardID, "CreatedAt": "2026-07-21T09:00:00Z",
				"Name": "Open", "Position": 1024, "Status": "OPEN",
				"UpdatedAt": "2026-07-21T09:00:00Z"}},
		{"board_tasks", "41000000-0000-4000-8000-000000000307", "TASK", taskID,
			map[string]any{"BoardID": boardID, "CancelledAt": nil, "ColumnID": columnID,
				"CompletedAt": nil, "CreatedAt": "2026-07-21T09:00:00Z",
				"CreatorID": creatorID, "DeletedAt": nil, "Description": "Legacy recurrence",
				"DueAt": "2026-07-22T09:00:00Z", "FormFilter": map[string]any{"status": "open"},
				"IsRecurring": true, "Kind": "CUSTOM", "Position": 1024,
				"Priority": "HIGH", "RecurrenceEnd": "2026-08-21T09:00:00Z",
				"RecurrenceRule": "FREQ=WEEKLY", "StartsAt": "2026-07-21T09:00:00Z",
				"Status": "OPEN", "Title": "Protected task",
				"UpdatedAt": "2026-07-21T09:00:00Z"}},
	}
	changes := make([]any, 0, len(records))
	for index, value := range records {
		key := map[string]any{"source_id": value.sourceID, "source_kind": value.sourceKind,
			"target_id": value.targetID, "target_kind": value.targetKind}
		after := map[string]any{"source_id": value.sourceID, "source_kind": value.sourceKind,
			"source_sha256": fmt.Sprintf("%064x", 60+index), "target_id": value.targetID,
			"target_kind": value.targetKind, "target_record": value.record}
		changes = append(changes, map[string]any{"after": after, "before": nil,
			"key": key, "kind": "insert", "ordinal": 1})
	}
	return []protectedWorkBatch{
		newProtectedWorkBatch(a, "task_boards", 1, "4101", a.finalLSN, []any{changes[0]}),
		newProtectedWorkBatch(a, "task_board_columns", 2, "4101", a.finalLSN, []any{changes[1]}),
		newProtectedQuietBatch(a, "board_members", 3, a.finalLSN),
		newProtectedWorkBatch(a, "board_tasks", 4, "4101", a.finalLSN, []any{changes[2]}),
		newProtectedQuietBatch(a, "board_task_comments", 5, a.finalLSN),
		newProtectedQuietBatch(a, "calendar_events", 6, a.finalLSN),
	}
}

func assertProtectedTaskMetadata(t *testing.T, ctx context.Context,
	database *sql.DB, schema string,
) {
	t.Helper()
	workInSchema(t, ctx, database, schema, func(transaction *sql.Tx) {
		var creator, priority, startsAt, rule, filter string
		var recurring bool
		var recurrenceEnd time.Time
		err := transaction.QueryRowContext(ctx, `
SELECT task.creator_user_id::text,task.priority,task.starts_at::text,
       metadata.is_recurring,metadata.recurrence_rule,metadata.recurrence_end,
       metadata.form_filter::text
FROM work_tasks AS task
JOIN work_task_compatibility_metadata AS metadata ON metadata.task_id=task.id
WHERE task.id='41000000-0000-4000-8000-000000000306'`).Scan(
			&creator, &priority, &startsAt, &recurring, &rule, &recurrenceEnd, &filter)
		expectedEnd := time.Date(2026, 8, 21, 9, 0, 0, 0, time.UTC)
		if err != nil || creator != "41000000-0000-4000-8000-000000000303" ||
			priority != "HIGH" || startsAt == "" || !recurring || rule != "FREQ=WEEKLY" ||
			!recurrenceEnd.Equal(expectedEnd) || filter != `{"status": "open"}` {
			t.Fatalf("protected task metadata creator=%s priority=%s starts=%s recurring=%t rule=%s end=%s filter=%s err=%v",
				creator, priority, startsAt, recurring, rule, recurrenceEnd, filter, err)
		}
	})
}
