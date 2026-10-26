package migrator

import (
	"strings"
	"testing"
)

func TestFormsCompletionTriggerAndBackfillAreDurable(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var triggerMigration string
	var backfillMigration string
	for _, document := range documents {
		if strings.Contains(document.Path(),
			"/fresh/20260718020023_enqueue_forms_schema_completion_v1") {
			triggerMigration = string(document.Content())
		}
		if strings.Contains(document.Path(),
			"/fresh/20260718020024_backfill_forms_schema_completion_v1") {
			backfillMigration = string(document.Content())
		}
	}
	if triggerMigration == "" || backfillMigration == "" {
		t.Fatal("Forms schema completion migrations are missing")
	}
	for _, token := range []string{
		"AFTER UPDATE ON platform.schema_jobs",
		"OLD.state NOT IN ('succeeded', 'failed', 'cancelled')",
		"NEW.state IN ('succeeded', 'failed', 'cancelled')",
		"NEW.submission_authority_id",
		"NEW.submission_authority_group_id",
		"NEW.submission_group_generation",
		"NEW.submission_billing_generation",
		"NEW.submission_forms_generation",
		"platform.platform_async_enqueue_forms_schema_completion_v1(",
		"ENABLE ALWAYS TRIGGER schema_jobs_forms_completion_v1",
	} {
		if !strings.Contains(triggerMigration, token) {
			t.Fatalf("Forms schema completion migration misses %q", token)
		}
	}
	for _, token := range []string{
		"platform.platform_migrator_backfill_forms_completion_v1",
		"forms-schema-completion-backfill.v1", "REVOKE ALL ON FUNCTION",
		"platform.platform_async_enqueue_forms_schema_completion_v1(",
	} {
		if !strings.Contains(backfillMigration, token) {
			t.Fatalf("Forms schema completion backfill misses %q", token)
		}
	}
	if strings.Contains(triggerMigration+backfillMigration, "TO harpy_runtime") {
		t.Fatal("migrator completion routines must not be runtime-callable")
	}
}
