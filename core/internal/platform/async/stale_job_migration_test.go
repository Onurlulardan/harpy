package async

import (
	"strings"
	"testing"
)

func TestStaleJobForwardMigrationsFreezeDiscoveryAndLeastPrivilege(t *testing.T) {
	t.Parallel()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	markers := []string{
		"CREATE OR REPLACE TRIGGER jobs_transition_guard",
		"CREATE FUNCTION platform.platform_async_guard_stale_job_park_v2()",
		"OLD.state NOT IN ('queued', 'running')",
		"CREATE FUNCTION platform.platform_async_park_stale_work_v2(",
		"v_job.effect_started_at IS NOT NULL",
		"harpy.job-stale-reference.v1",
		"reason_class, reason_code",
		"GRANT EXECUTE ON FUNCTION platform.platform_async_park_stale_work_v2(",
		"platform_async_validate_stale_external_attempt_unknown_v1(",
		"platform_async_validate_stale_external_job_unknown_v1(",
		"CREATE FUNCTION platform.platform_async_guard_job_attempt_v2()",
		"CREATE FUNCTION platform.platform_async_guard_job_transition_v2()",
		"CREATE FUNCTION platform.platform_async_finalize_stale_external_unknown_v1(",
		"NEW.error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN'",
		"ERRCODE = 'P5678'",
		"GRANT EXECUTE ON FUNCTION\n    platform.platform_async_finalize_stale_external_unknown_v1(",
	}
	for _, origin := range []string{"fresh", "v1_adopted"} {
		var source strings.Builder
		var finalizer string
		for _, document := range documents {
			if strings.Contains(document.Path(), "/"+origin+"/202607150217") {
				source.Write(document.Content())
				if strings.Contains(document.Path(), "20260715021750_") {
					finalizer = string(document.Content())
				}
			}
		}
		for _, marker := range markers {
			if !strings.Contains(source.String(), marker) {
				t.Fatalf("%s stale job migrations lack %q", origin, marker)
			}
		}
		if finalizer == "" || strings.Contains(finalizer, "INSERT INTO platform.dead_letters") {
			t.Fatalf("%s stale unknown finalizer duplicated terminal DLQ ownership", origin)
		}
		if strings.Contains(source.String(), "CREATE OR REPLACE FUNCTION") {
			t.Fatalf("%s stale migration replaced an applied routine", origin)
		}
	}
}
