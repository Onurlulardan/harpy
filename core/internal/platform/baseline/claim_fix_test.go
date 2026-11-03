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

func TestTenantCompletionUsesSchemaHistoryAndReopensUnactivatedJobs(t *testing.T) {
	for _, table := range []string{
		"platform.release_generations",
		"platform.schema_contracts",
		"platform.migration_runs",
		"platform.tenant_migration_requirement_catalogs",
		"platform.tenant_migration_checkpoints",
		"platform_migrator_verify_new_tenant_v1",
	} {
		if strings.Contains(platformTenantExecutorCompletionScript, table) {
			t.Fatalf("tenant completion baseline still reads %s", table)
		}
	}
	for _, required := range []string{
		"tenancy_tenant_migration_reopen_unactivated_v1",
		"goose_db_version",
		"20261028000001",
		"20261102000001",
		"DROP SCHEMA IF EXISTS %I CASCADE",
		"SET state = 'pending'",
		"PERFORM platform.tenancy_tenant_migration_reopen_unactivated_v1()",
	} {
		if !strings.Contains(platformTenantExecutorCompletionScript, required) {
			t.Fatalf("tenant completion baseline is missing %s", required)
		}
	}
}
