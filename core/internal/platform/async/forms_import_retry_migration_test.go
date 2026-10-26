package async

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func TestFormsImportRetryMigrationIsSafeAndPaired(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	logical := []string{
		"20260801015000_create_forms_import_retry_v1",
		"20260801015001_allow_forms_import_retry_read_v1",
		"20260801015002_allow_forms_import_retry_control_v1",
		"20260801015003_allow_forms_import_retry_current_lease_v1",
		"20260801015004_allow_forms_import_retry_complete_v1",
	}
	pairs := make(map[string][]byte, len(logical))
	for _, name := range logical {
		pairs[name] = formsImportRetryMigrationPair(t, documents, name)
	}
	sql := string(pairs[logical[0]])
	for _, required := range []string{
		"platform_async_forms_import_retry_v1",
		"pg_catalog.uuid_extract_version(p_retry_job_id)",
		"forms.records.import.resume",
		"v_source.state IS DISTINCT FROM 'dead'",
		"v_source.cancel_requested_at IS NOT NULL",
		"v_source.effect_started_at IS NOT NULL",
		"platform_async_assert_current_authority_v1",
		"v_retry.retry_of_job_id<>p_source_job_id",
		"'forms.record-import.retry:' || v_source.id::text",
		"ON CONFLICT ON CONSTRAINT jobs_dedupe_unique DO NOTHING",
		"retry_of_job_id",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("Forms import retry migration misses %q", required)
		}
	}
	if strings.Contains(sql, "GRANT EXECUTE") {
		t.Fatal("platform retry primitive must remain private to owner wrappers")
	}
	if strings.Count(sql, "platform_async_assert_current_authority_v1") != 2 {
		t.Fatal("new retry and replay must both assert current authority")
	}
	for _, name := range logical[1:] {
		sql = string(pairs[name])
		for _, required := range []string{
			"DROP FUNCTION platform.platform_async_forms_import_",
			"CREATE FUNCTION platform.platform_async_forms_import_",
			"job.retry_of_job_id IS NULL",
			"'forms.record-import:' || p_import_id::text",
			"job.retry_of_job_id IS NOT NULL",
			"'forms.record-import.retry:' || job.retry_of_job_id::text",
			"TO harpy_runtime",
		} {
			if !strings.Contains(sql, required) {
				t.Fatalf("%s misses retry lineage guard %q", name, required)
			}
		}
	}
}

func formsImportRetryMigrationPair(
	t *testing.T, documents []migration.Document, logical string,
) []byte {
	t.Helper()
	paired := map[string][]byte{}
	for _, document := range documents {
		for _, origin := range []string{"fresh", "v1_adopted"} {
			if strings.HasSuffix(document.Path(), logical+"."+origin+".sql") {
				paired[origin] = document.Content()
			}
		}
	}
	if len(paired) != 2 || !bytes.Equal(paired["fresh"], paired["v1_adopted"]) {
		t.Fatalf("%s migration pair is missing or differs", logical)
	}
	return paired["fresh"]
}
