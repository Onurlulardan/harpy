package cutover

import (
	"strings"
	"testing"
)

func TestGen70ActivationCapabilityFenceIsPairedAndForwardOnly(t *testing.T) {
	const version = "20260729100034"
	pair := successorFinalizationMigrationPair(t, version)
	if len(pair) != 2 || pair["fresh"] != pair["v1_adopted"] {
		t.Fatal("Gen70 activation capability fence is not paired byte-identically")
	}
	source := pair["fresh"]
	if strings.Count(source, "-- +goose Up") != 1 ||
		strings.Contains(source, "-- +goose Down") ||
		strings.Contains(source, "CREATE OR REPLACE") ||
		strings.Contains(source, "DO $") {
		t.Fatal("Gen70 activation capability fence is not forward-only")
	}
}

func TestGen70ActivationCapabilityFenceCoversPlatformContracts(t *testing.T) {
	source := successorFinalizationMigrationPair(
		t, "20260729100034",
	)["fresh"]
	for _, marker := range []string{
		"platform.tenancy_onboarding_activate_v3(uuid,text,bigint)",
		"platform.tenancy_onboarding_activate_v2(uuid,text,bigint)",
		"platform.tenancy_activation_initialize_v5(",
		"platform.tenancy_authority_reconcile_current_v1(",
		"platform.billing_installation_assign_v3(",
		"platform.iam_installation_activation_assert_v2(",
		"platform.iam_installation_security_epoch_initialize_v2(",
		"platform.iam_installation_identity_directory_upsert_v1(",
		"platform.iam_installation_instance_owner_bind_v2(",
		"platform.billing_installation_open_usage_v2(",
		"platform.platformops_installation_recovery_initialize_v2(",
		"platform.platform_async_initial_admin_setup_enqueue_v2(",
		"platform.platform_async_installation_activation_event_v2(",
		"platform.platformauthority_installation_owner_bind_v3(",
		"platform.tenancy_onboarding_complete_v2(",
		"platform.platformauthority_installation_head_pin_v3(",
		"platform.tenancy_onboarding_accept_v2(",
		"routine.proowner",
		"routine.prosecdef",
		"search_path=pg_catalog",
		"'harpy_runtime', routine_oid, 'EXECUTE'",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Gen70 public activation fence lacks %q", marker)
		}
	}
}

func TestGen70ActivationCapabilityFenceClosesInternalStorage(t *testing.T) {
	source := successorFinalizationMigrationPair(
		t, "20260729100034",
	)["fresh"]
	for _, marker := range []string{
		"platform.tenancy_activation_step_assert_v1(",
		"platform.tenancy_activation_step_consume_v1(",
		"platform.tenancy_activation_step_context_assert_v1(",
		"platform.tenancy_activation_step_commit_guard_v1()",
		"platform.platform_async_installation_activation_evidence_assert_v1(",
		"platform.tenancy_activation_command_consume_v1(",
		"platform.tenancy_authority_reconcile_scope_assert_v1(",
		"platform.tenancy_activation_capability_commit_guard_v1()",
		"platform.tenancy_authority_scope_commit_guard_v1()",
		"platform.billing_installation_id_v1(",
		"platform.analytics_reconcile_current_tenant_authorities_v2()",
		"platform.workflow_reconcile_current_tenant_authorities_v2()",
		"platform.billing_reconcile_current_global_authority_v2()",
		"platform.billing_reconcile_current_tenant_authorities_v2()",
		"platform.tenancy_activation_step_capabilities",
		"platform.platform_async_installation_activation_receipts",
		"platform.tenancy_activation_command_capabilities",
		"platform.tenancy_authority_reconcile_scopes",
		"platform.billing_installation_assignment_contexts",
		"VALUES ('harpy_runtime'), ('harpy_v2_runtime')",
		"('SELECT'), ('INSERT'), ('UPDATE'), ('DELETE')",
		"relation.relowner",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Gen70 internal activation fence lacks %q", marker)
		}
	}
}

func TestGen70ActivationCapabilityFencePreservesGen23ABI(t *testing.T) {
	source := successorFinalizationMigrationPair(
		t, "20260729100034",
	)["fresh"]
	for _, marker := range []string{
		"platform.tenancy_onboarding_activate_v1(uuid,text,bigint)",
		"platform.billing_installation_assign_v1(",
		"platform.billing_installation_assign_v2(",
		"platform.billing_reconcile_current_global_authority_v1()",
		"platform.billing_reconcile_current_tenant_authorities_v1()",
		"platform.billing_register_global_authority_v1(",
		"platform.billing_register_tenant_authority_v1(",
		"platform.platformauthority_installation_head_pin_v2(",
		"platform.tenancy_activation_initialize_v4(",
		"platform.tenancy_onboarding_complete_v1(",
		"platform.tenancy_onboarding_accept_v1(",
		"platform.billing_installation_open_usage_v1(",
		"platform.platformops_installation_recovery_initialize_v1(",
		"platform.platform_async_initial_admin_setup_enqueue_v1(",
		"platform.platform_async_installation_activation_event_v1(",
		"platform.platformauthority_installation_owner_bind_v2(",
		"platform.tenancy_activation_assert_v1(",
		"platform.tenancy_activation_assert_v2(",
		"platform.iam_security_epoch_initialize_v1(",
		"platform.iam_installation_instance_owner_bind_v1(",
		"platform.iam_installation_instance_owner_input_v1(",
		"legacy_resolved",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Gen70 compatibility fence lacks %q", marker)
		}
	}
}

func TestGen71ActivationContractIsPairedAndForwardOnly(t *testing.T) {
	const version = "20260729100045"
	pair := successorFinalizationMigrationPair(t, version)
	if len(pair) != 2 || pair["fresh"] != pair["v1_adopted"] {
		t.Fatal("Gen71 activation contract is not paired byte-identically")
	}
	source := pair["fresh"]
	if strings.Count(source, "-- +goose Up") != 1 ||
		strings.Contains(source, "-- +goose Down") ||
		strings.Contains(source, "CREATE OR REPLACE") ||
		strings.Contains(source, "DROP FUNCTION") ||
		strings.Contains(source, "DO $") {
		t.Fatal("Gen71 activation contract is not forward-only")
	}
}

func TestGen71ActivationContractRetiresLegacyAndPinsSuccessors(t *testing.T) {
	source := successorFinalizationMigrationPair(
		t, "20260729100045",
	)["fresh"]
	for _, marker := range []string{
		"platform.tenancy_onboarding_activate_v2(uuid,text,bigint)",
		"platform.tenancy_activation_initialize_v4(",
		"platform.iam_installation_instance_owner_bind_v1(",
		"platform.billing_installation_assign_v2(",
		"platform.analytics_register_tenant_authority_v1(",
		"platform.workflow_register_tenant_authority_v1(",
		"platform.platformops_installation_recovery_initialize_v1(",
		"platform.platform_async_initial_admin_setup_enqueue_v1(",
		"platform.platformauthority_installation_owner_bind_v2(",
		"platform.platformauthority_installation_head_pin_v2(",
		"retired_resolved CROSS JOIN forbidden_roles",
		"platform.tenancy_onboarding_activate_v3(uuid,text,bigint)",
		"platform.tenancy_activation_assert_v2(uuid,uuid,name)",
		"platform.iam_installation_instance_owner_input_v1(uuid)",
		"platform.billing_installation_assign_v3(",
		"platform.platform_async_initial_admin_setup_enqueue_v2(",
		"platform.platformauthority_installation_head_pin_v3(",
		"platform.analytics_reconcile_current_tenant_authorities_v2()",
		"platform.workflow_reconcile_current_tenant_authorities_v2()",
		"platform.tenancy_activation_step_capabilities",
		"platform.platform_async_installation_activation_receipts",
		"platform.billing_installation_assignment_contexts",
		"'harpy_v2_runtime', routine_oid, 'EXECUTE'",
		"pg_catalog.has_table_privilege(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Gen71 activation contract lacks %q", marker)
		}
	}
}
