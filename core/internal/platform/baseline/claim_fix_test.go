package baseline

import (
	"strings"
	"testing"
)

func TestTenantExecutorBaselineDoesNotReadRemovedReleaseEvidence(t *testing.T) {
	for _, table := range []string{
		"platform.release_generations",
		"platform.schema_contracts",
		"platform.migration_runs",
		"platform.bootstrap_receipts",
		"platform.extension_registry",
		"platform.database_generation_manifests",
	} {
		if strings.Contains(platformTenantExecutorBaselineScript, table) {
			t.Fatalf("tenant executor baseline still reads %s", table)
		}
	}
	if !strings.Contains(platformTenantExecutorBaselineScript, "IF NOT EXISTS (SELECT 1 FROM platform.tenancy_deployment_policy) THEN") ||
		!strings.Contains(platformTenantExecutorBaselineScript, "tenant migration deployment policy conflicts") ||
		!strings.Contains(platformTenantExecutorBaselineScript, "tenant migration readiness is stale") {
		t.Fatal("tenant executor baseline dropped the idle and conflict checks")
	}
}
