package postgres

import (
	"strings"
	"testing"
)

func TestRetentionSQLUsesOnlyTheArglessCanonicalCycle(t *testing.T) {
	for _, marker := range []string{
		"platform.platform_async_run_retention_cycle_v1()",
		"lock_acquired", "jobs_redacted", "dead_letters_purged",
	} {
		if !strings.Contains(runRetentionCycleSQL, marker) {
			t.Fatalf("retention SQL lacks %q", marker)
		}
	}
	if strings.Contains(runRetentionCycleSQL, "$1") ||
		strings.Contains(runRetentionCycleSQL, "_payloads_v") ||
		strings.Contains(runRetentionCycleSQL, "purge_") {
		t.Fatal("retention SQL bypassed the sealed cycle wrapper")
	}
}
