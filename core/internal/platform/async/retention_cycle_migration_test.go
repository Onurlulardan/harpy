package async

import (
	"strings"
	"testing"
)

func TestRetentionCycleMigrationIsPairedSealedAndLeastPrivilege(t *testing.T) {
	t.Parallel()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	const logicalName = "20260715022000_create_async_retention_cycle_v1"
	var fresh []byte
	for _, document := range documents {
		if strings.HasSuffix(document.Path(), logicalName+".fresh.sql") {
			fresh = document.Content()
		}
	}
	if len(fresh) == 0 {
		t.Fatal("fresh retention cycle migration did not load")
	}
	source := string(fresh)
	markers := []string{
		"platform_async_run_retention_cycle_v1()", "RETURNS TABLE (",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"pg_catalog.pg_try_advisory_xact_lock(", "SET search_path = pg_catalog",
		"platform_async_redact_job_payloads_v1(500)",
		"platform_async_redact_outbox_payloads_v2(500)",
		"platform_async_redact_dead_letter_payloads_v2(500)",
		"platform_async_purge_inbox_v1(500)",
		"platform_async_purge_idempotency_receipts_v1(500)",
		"platform_async_purge_jobs_v1(500)",
		"platform_async_purge_dead_letters_v2(500)",
	}
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("retention cycle migration lacks %q", marker)
		}
	}
	assertRetentionStepOrder(t, source, markers[6:13])
	assertRetentionRoutineACL(t, source)
	for _, forbidden := range []string{
		"platform_async_redact_outbox_payloads_v1(500)",
		"platform_async_redact_dead_letter_payloads_v1(500)",
		"platform_async_purge_dead_letters_v1(500)",
		"p_batch_size", "platform_async_run_retention_cycle_v1(integer)",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("retention cycle migration contains %q", forbidden)
		}
	}
}

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
