package baseline

import (
	"strings"
	"testing"
)

func TestTenantExecutorClaimFixIdlesWithoutReleaseGraph(t *testing.T) {
	script := platformTenantExecutorClaimFixScript
	idle := "IF NOT EXISTS (SELECT 1 FROM platform.tenancy_deployment_policy) THEN"
	idleAt := strings.Index(script, idle)
	joinAt := strings.Index(script, "JOIN platform.release_generations")
	if idleAt < 0 || joinAt < 0 || idleAt > joinAt || !strings.Contains(script[idleAt:joinAt], "RETURN;") {
		t.Fatal("claim fix must return before the release-graph join when no deployment policy exists")
	}
	if !strings.Contains(script, "NOT EXISTS (SELECT 1 FROM platform.release_generations)") ||
		!strings.Contains(script, "NOT EXISTS (SELECT 1 FROM platform.schema_contracts)") {
		t.Fatal("heartbeat fix must refresh readiness when the release graph is absent")
	}
	for _, message := range []string{
		"tenant migration executor is not claim ready",
		"tenant migration deployment policy conflicts",
		"tenant migration readiness is stale",
	} {
		if !strings.Contains(script, message) {
			t.Fatalf("claim fix dropped fail-closed error %q", message)
		}
	}
}
