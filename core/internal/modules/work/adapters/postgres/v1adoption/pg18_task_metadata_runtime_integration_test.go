//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

const runtimeMetadataTaskID = "31000000-0000-4000-8000-000000000012"

func TestWorkPG18RuntimeCompatibilityMetadataRoundTripAndReader(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := startWorkPG18(t, ctx)
	prepareWorkSchemas(t, ctx, database.owner)
	seedWorkQueryFixture(t, ctx, database.owner, "work_fresh_a")
	prepareWorkQueryRuntime(t, ctx, database.owner)
	connection, err := database.owner.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if _, err = connection.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_runtime;
SET search_path TO work_fresh_a,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	assertWorkQueryRuntimeBoundary(t, ctx, connection)
	assertRuntimeMetadataACL(t, ctx, connection)
	assertRuntimeMetadataRoundTrip(t, ctx, connection)
	assertRuntimeCompatibilityReadSQL(t, ctx, connection)
}

func assertRuntimeMetadataACL(t *testing.T, ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	var selectTable, insertTask, updateFilter, updateTask, deleteTable bool
	err := connection.QueryRowContext(ctx, `
SELECT has_table_privilege(current_user,'work_task_compatibility_metadata','SELECT'),
       has_column_privilege(current_user,'work_task_compatibility_metadata','task_id','INSERT'),
       has_column_privilege(current_user,'work_task_compatibility_metadata','form_filter','UPDATE'),
       has_column_privilege(current_user,'work_task_compatibility_metadata','task_id','UPDATE'),
       has_table_privilege(current_user,'work_task_compatibility_metadata','DELETE')`).Scan(
		&selectTable, &insertTask, &updateFilter, &updateTask, &deleteTable)
	if err != nil || !selectTable || !insertTask || !updateFilter ||
		updateTask || deleteTable {
		t.Fatalf("metadata ACL select=%t insert_task=%t update_filter=%t "+
			"update_task=%t delete=%t err=%v", selectTable, insertTask,
			updateFilter, updateTask, deleteTable, err)
	}
}

func assertRuntimeMetadataRoundTrip(t *testing.T, ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	upsert := loadWorkAdapterSQL(t, "upsert_legacy_task_metadata.sql")
	load := loadWorkAdapterSQL(t, "load_legacy_task_metadata_for_update.sql")
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	end := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	if _, err = transaction.ExecContext(ctx, upsert, runtimeMetadataTaskID, true,
		"FREQ=WEEKLY", end, []byte(`{"amount":{"gt":100},"status":"open"}`)); err != nil {
		t.Fatal(err)
	}
	assertLoadedRuntimeMetadata(t, ctx, transaction, load, true,
		"FREQ=WEEKLY", &end, `{"amount":{"gt":100},"status":"open"}`)
	if _, err = transaction.ExecContext(ctx, upsert, runtimeMetadataTaskID, false,
		"FREQ=DAILY", nil, []byte(`{"priority":2,"status":"closed"}`)); err != nil {
		t.Fatal(err)
	}
	assertLoadedRuntimeMetadata(t, ctx, transaction, load, false,
		"FREQ=DAILY", nil, `{"priority":2,"status":"closed"}`)
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertLoadedRuntimeMetadata(t *testing.T, ctx context.Context,
	transaction *sql.Tx, query string, recurring bool, rule string,
	expectedEnd *time.Time, filter string,
) {
	t.Helper()
	var actualRecurring bool
	var actualRule string
	var end *time.Time
	var actualFilter []byte
	err := transaction.QueryRowContext(ctx, query, runtimeMetadataTaskID).Scan(
		&actualRecurring, &actualRule, &end, &actualFilter)
	if err != nil || actualRecurring != recurring || actualRule != rule ||
		!sameRuntimeMetadataTime(end, expectedEnd) ||
		!jsonObjectsEqual(actualFilter, []byte(filter)) {
		t.Fatalf("metadata recurring=%t rule=%s end=%v filter=%s err=%v",
			actualRecurring, actualRule, end, actualFilter, err)
	}
}

func sameRuntimeMetadataTime(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func assertRuntimeCompatibilityReadSQL(t *testing.T, ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	query := loadWorkAdapterSQL(t, "read_visible_records.sql")
	zero := uuid.Nil
	var document []byte
	err := connection.QueryRowContext(ctx, query, "tasks",
		uuid.MustParse("31000000-0000-4000-8000-000000000001"), []string{"USER"},
		[]uuid.UUID{uuid.MustParse("31000000-0000-4000-8000-000000000001")},
		uuid.MustParse(runtimeMetadataTaskID), zero, 10, "", zero, nil, nil,
		false, false, "", true).Scan(&document)
	if err != nil {
		t.Fatal(err)
	}
	var records []map[string]any
	if json.Unmarshal(document, &records) != nil || len(records) != 1 ||
		records[0]["recurrence_rule"] != "FREQ=DAILY" ||
		records[0]["is_recurring"] != false {
		t.Fatalf("compatibility Reader SQL result=%s", document)
	}
	filter, ok := records[0]["form_filter"].(map[string]any)
	if !ok || filter["status"] != "closed" || filter["priority"] != float64(2) {
		t.Fatalf("compatibility filter result=%s", document)
	}
}

func loadWorkAdapterSQL(t *testing.T, name string) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workingDirectory, "..", "sql", "tenant", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func jsonObjectsEqual(left, right []byte) bool {
	var leftValue, rightValue map[string]any
	return json.Unmarshal(left, &leftValue) == nil &&
		json.Unmarshal(right, &rightValue) == nil &&
		jsonEqual(leftValue, rightValue)
}

func jsonEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}
