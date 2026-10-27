package async

import (
	"strings"
	"testing"
)

func TestFormsCompletionSubmissionIsFrozenAndPrivate(t *testing.T) {
	migrations := freshAsyncMigrations(t)
	assertion := migrations["20260718020020_create_forms_schema_completion_assertion_v1"]
	for _, token := range []string{
		"platform.platform_async_assert_forms_schema_completion_job_v1",
		"submission_authority_id", "submission_authority_group_id",
		"submission_group_generation", "submission_billing_generation",
		"submission_forms_generation", "SESSION_USER IS DISTINCT FROM 'harpy_migrator'",
	} {
		if !strings.Contains(assertion, token) {
			t.Fatalf("Forms completion assertion misses %q", token)
		}
	}
	enqueue := migrations["20260718020021_create_forms_schema_completion_enqueue_v1"]
	for _, token := range []string{
		"platform.platform_async_enqueue_forms_schema_completion_v1",
		"platform.billing_resolve_current_tenant_authority_receipt_v1",
		"forms-publication-apply-result/", "forms.publication.apply-result",
		"database_transactional", "ON CONFLICT ON CONSTRAINT jobs_dedupe_unique",
		"REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(enqueue, token) {
			t.Fatalf("Forms completion enqueue misses %q", token)
		}
	}
	if strings.Contains(enqueue, "GRANT EXECUTE") {
		t.Fatal("private Forms completion enqueue helper is runtime-callable")
	}
	guard := migrations["20260718020022_replace_async_owner_insert_guard_v3"]
	if !strings.Contains(guard,
		"platform.platform_async_assert_forms_schema_completion_job_v1(NEW)") {
		t.Fatal("async owner guard does not bind exact Forms completion insertion")
	}
}

func freshAsyncMigrations(t *testing.T) map[string]string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]string{}
	for _, document := range documents {
		if !strings.Contains(document.Path(), "/fresh/") {
			continue
		}
		name := document.Path()[strings.LastIndex(document.Path(), "/")+1:]
		result[strings.TrimSuffix(name, ".fresh.sql")] = string(document.Content())
	}
	return result
}
