package postgres

import (
	"strings"
	"testing"
)

func TestLayoutInstallationCapabilityFenceIsPairedAndComplete(t *testing.T) {
	const name = "20260729100036_assert_installation_capabilities_v1"
	fresh := readLayoutMigration(t, name)
	assertLayoutMigrationSource(t, fresh)
	source := string(fresh)
	for _, marker := range []string{
		"system_resource_layout_installation_assert_v2(",
		"system_resource_layout_installation_record_v2(",
		"platform.tenancy_activation_step_context_assert_v1(",
		"platform.tenancy_activation_step_consume_v1(",
		"system_resource_layout_installation_record_v1(",
		"routine.proowner",
		"routine.prosecdef",
		"current_setting('search_path')",
		"'harpy_runtime', routine_oid, 'EXECUTE'",
		"'harpy_v2_runtime', routine_oid, 'EXECUTE'",
		"SELECT 1 / CASE WHEN",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("layout installation capability fence lacks %q", marker)
		}
	}
}

func TestLayoutInstallationAuthorityContractIsPairedAndComplete(t *testing.T) {
	const name = "20260729100047_contract_system_resource_layout_installation_authority_v1"
	fresh := readLayoutMigration(t, name)
	assertLayoutMigrationSource(t, fresh)
	source := string(fresh)
	for _, marker := range []string{
		"system_resource_layout_installation_record_v1(uuid,uuid,uuid,uuid,uuid,timestamptz)",
		"system_resource_layout_installation_assert_v2",
		"system_resource_layout_installation_record_v2",
		"system_resource_layout_installation_baseline_guard_v1",
		"platform.tenancy_activation_step_context_assert_v1",
		"platform.tenancy_activation_step_consume_v1",
		"retired_resolved",
		"retained_resolved",
		"internal_resolved",
		"routine_oid IS NULL",
		"routine.proowner",
		"routine.prosecdef",
		"'harpy_runtime', routine_oid, 'EXECUTE'",
		"'harpy_v2_runtime', routine_oid, 'EXECUTE'",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("layout authority contract lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"-- +goose Down", "CREATE FUNCTION", "CREATE OR REPLACE", "DROP ",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("layout authority contract contains %q", forbidden)
		}
	}
}
