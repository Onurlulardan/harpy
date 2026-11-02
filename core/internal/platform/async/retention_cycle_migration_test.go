package async

import (
	"strings"
	"testing"
)

func assertRetentionRoutineACL(t *testing.T, source string) {
	t.Helper()
	const wrapperGrant = "GRANT EXECUTE ON FUNCTION\n" +
		"    platform.platform_async_run_retention_cycle_v1() TO harpy_runtime;"
	if strings.Count(source, wrapperGrant) != 1 {
		t.Fatal("retention wrapper exact runtime grant drifted")
	}
	leaves := []string{
		"platform.platform_async_redact_job_payloads_v1(integer)",
		"platform.platform_async_redact_outbox_payloads_v2(integer)",
		"platform.platform_async_redact_dead_letter_payloads_v2(integer)",
		"platform.platform_async_purge_inbox_v1(integer)",
		"platform.platform_async_purge_idempotency_receipts_v1(integer)",
		"platform.platform_async_purge_jobs_v1(integer)",
		"platform.platform_async_purge_dead_letters_v2(integer)",
	}
	for _, leaf := range leaves {
		revoke := "REVOKE EXECUTE ON FUNCTION\n    " + leaf +
			"\nFROM harpy_runtime;"
		if strings.Count(source, revoke) != 1 {
			t.Fatalf("retention leaf exact runtime revoke drifted: %s", leaf)
		}
	}
	if strings.Count(source, "REVOKE EXECUTE ON FUNCTION") != len(leaves) {
		t.Fatal("retention leaf runtime revoke set is not exact")
	}
}

func assertRetentionStepOrder(t *testing.T, source string, steps []string) {
	t.Helper()
	previous := -1
	for _, step := range steps {
		index := strings.Index(source, step)
		if index <= previous {
			t.Fatalf("retention step order drifted at %q", step)
		}
		previous = index
	}
}
