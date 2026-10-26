package postgres

import (
	"strings"
	"testing"
)

func TestLegacyTaskMetadataSQLIsTaskScopedAndBounded(t *testing.T) {
	for name, statement := range map[string]string{
		"load":   loadLegacyTaskMetadataForUpdateSQL,
		"upsert": upsertLegacyTaskMetadataSQL,
	} {
		if !strings.Contains(statement, "work_task_compatibility_metadata") ||
			!strings.Contains(statement, "task_id") {
			t.Fatalf("%s metadata SQL lost task scope: %s", name, statement)
		}
	}
	if !strings.Contains(loadLegacyTaskMetadataForUpdateSQL, "WHERE task.id = $1") ||
		!strings.Contains(loadLegacyTaskMetadataForUpdateSQL, "FOR UPDATE OF task") {
		t.Fatal("metadata load is not an exact locked task read")
	}
	if !strings.Contains(upsertLegacyTaskMetadataSQL, "ON CONFLICT (task_id)") ||
		strings.Contains(upsertLegacyTaskMetadataSQL, "DELETE") {
		t.Fatal("metadata persistence is not an idempotent task upsert")
	}
}
