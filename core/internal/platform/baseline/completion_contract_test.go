package baseline

import (
	"strings"
	"testing"
)

func TestCompletionContractUsesGooseFloorAndOnboardingTransition(t *testing.T) {
	script := platformTenantExecutorCompletionContractScript
	for _, removed := range []string{
		"20261028000001",
		"platform.tenant_migration_checkpoints",
		"platform.tenant_migration_requirement_catalogs",
		"platform.migration_runs",
		"platform_migrator_verify_new_tenant_v1",
		"ARRAY[20261028000001, 20261102000001]",
	} {
		if strings.Contains(script, removed) {
			t.Fatalf("completion contract still embeds %s", removed)
		}
	}
	for _, required := range []string{
		"tenancy_tenant_migration_complete_v3",
		"minimum_schema_version CONSTANT bigint := 20261102000001",
		"tenant schema goose history is below the minimum schema version",
		"v_job.state NOT IN ('running', 'succeeded')",
		"receipt_kind = 'migration_completed'",
		"'authority_pending'",
		"'migration_ready'",
		"tenant migration onboarding transition is invalid",
		"SET state = v_next_state",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("completion contract is missing %s", required)
		}
	}
}
