package postgres

import (
	"strings"
	"testing"
)

func TestReaderSQLCarriesTypedTaskAndCompatibilityProjection(t *testing.T) {
	for _, fragment := range []string{
		"task.priority = $14", "CASE WHEN $15", "work_task_labels",
		"work_checklist_items", "work_task_assignees", "creator_user_id",
		"work_task_compatibility_metadata", "metadata.form_filter",
		"metadata.is_recurring", "metadata.recurrence_rule",
	} {
		if !strings.Contains(readVisibleRecordsSQL, fragment) {
			t.Fatalf("read query missing %q", fragment)
		}
	}
	if strings.Count(readVisibleRecordsSQL, "'form_slug'") != 1 ||
		strings.Count(readVisibleRecordsSQL, "'form_data_id'") != 1 {
		t.Fatal("task form aliases must be owner-enriched; only calendar defaults remain")
	}
}
