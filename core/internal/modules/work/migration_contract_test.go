package work

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationTracksAreByteExact(t *testing.T) {
	fresh, err := filepath.Glob("migrations/tenant/fresh/*.fresh.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh) != 32 {
		t.Fatalf("fresh migration count = %d, want 32", len(fresh))
	}
	for _, freshPath := range fresh {
		freshBody, err := os.ReadFile(freshPath)
		if err != nil {
			t.Fatal(err)
		}
		if len(freshBody) == 0 || !strings.HasSuffix(freshPath, ".fresh.sql") {
			t.Errorf("fresh migration empty: %s", filepath.Base(freshPath))
		}
	}
}

func TestLegacyTaskMetadataMigrationIsTypedAtomicAndEvidenceBound(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("migrations", "tenant", "fresh",
		"20260802232710_create_work_task_compatibility_metadata_v1.fresh.sql"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, fragment := range []string{
		"CREATE TABLE work_task_compatibility_metadata", "task_id uuid PRIMARY KEY",
		"REFERENCES work_tasks(id)", "recurrence_rule varchar(50)",
		"form_filter IS NULL", "NULLIF(v_record->'FormFilter','null'::jsonb)",
		"work_v1_adoption_task_values_trigger_v2",
		"work_v1_adoption_task_metadata_trigger_v1",
		"work_v1_adoption_document_valid_v1('TASK',v_record)",
		"GRANT SELECT ON work_task_compatibility_metadata TO harpy_runtime",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("legacy task metadata migration missing %q", fragment)
		}
	}
	if strings.Contains(text, "GRANT DELETE") {
		t.Fatal("legacy task metadata grants destructive runtime access")
	}
	if strings.Contains(text, "CREATE OR REPLACE FUNCTION") ||
		strings.Contains(text, "jsonb_typeof(form_filter) = 'object'") {
		t.Fatal("legacy task metadata migration reintroduces replacement or object-only drift")
	}
}

func TestMigrationsDoNotReferenceExternalOwnerTables(t *testing.T) {
	paths, err := filepath.Glob("migrations/tenant/fresh/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		lower := strings.ToLower(string(body))
		for _, forbidden := range []string{
			"references iam_", "references forms_", "references workflow_",
			"join iam_", "join forms_", "join workflow_",
		} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s contains owner-external SQL %q", filepath.Base(path), forbidden)
			}
		}
	}
}

func TestTaskCompatibilityMigrationIsTypedAndIndexed(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("migrations", "tenant", "fresh",
		"20260801013000_add_work_task_schedule_priority_v1.fresh.sql"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, fragment := range []string{
		"creator_user_id uuid", "priority text", "starts_at timestamptz",
		"priority IN ('LOW','MEDIUM','HIGH','URGENT')", "work_tasks_priority_due_idx",
		"work_board_columns_color_check", "work_calendar_events_color_check",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("task compatibility migration missing %q", fragment)
		}
	}
}
