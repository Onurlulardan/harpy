package async

import (
	"bytes"
	"os"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestJobRetentionReleaseMigrationIsPairedAndFailClosed(t *testing.T) {
	const name = "20260715021250_release_async_job_retention_v1"
	fresh, err := os.ReadFile("migrations/fresh/" + name + ".fresh.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh) == 0 {
		t.Fatal("fresh job retention release migration did not load")
	}
	if bytes.Count(fresh, []byte("\n")) > 200 {
		t.Fatal("job retention release migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("inspect job retention release migration: %v", err)
	}
	source := string(fresh)
	markers := []string{
		"platform_async_release_job_retention_v1",
		"DROP CONSTRAINT jobs_time_check", "completed_at <= updated_at",
		"SECURITY DEFINER", "SET search_path = pg_catalog",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"platform_async_assert_current_authority_v1(",
		"job.retention_release_revision = p_expected_revision",
		"p_retain_until >= job.retain_until", "p_retain_until <= v_now",
		"dependency.async_job_id = job.id", "retry.retry_of_job_id = job.id",
		"attempt.finished_at IS NULL", "receipt.retain_until > v_now",
		"RETURNING true INTO v_released", "REVOKE ALL ON FUNCTION",
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("job retention release migration lacks %q", marker)
		}
	}
}
