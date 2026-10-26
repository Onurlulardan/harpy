package billing

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestBillingAuthorityReconcileMigrationsArePairedAndOwnerSafe(t *testing.T) {
	tests := []struct {
		name    string
		markers []string
	}{
		{
			name: "20260729100005_create_billing_global_authority_reconcile_v1",
			markers: []string{
				"billing_reconcile_current_global_authority_v1",
				"billing_register_global_authority_v1(",
				"harpy:billing:global-authority-reconcile:v1",
				"EXCEPTION WHEN SQLSTATE '23505'",
			},
		},
		{
			name: "20260729100006_create_billing_tenant_authority_reconcile_v1",
			markers: []string{
				"billing_reconcile_current_tenant_authorities_v1",
				"billing_register_tenant_authority_v1(",
				"harpy:billing:tenant-authority-reconcile:v1",
				"EXCEPTION WHEN SQLSTATE '23505'",
				"extra.ordinal = 2 AND extra.domain = 'forms'",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := pairedBillingAuthorityMigration(t, test.name)
			for _, marker := range append(test.markers,
				"harpy-v2:core-operator-initialize",
				"pg_catalog.uuidv7()",
				"TO harpy_runtime",
			) {
				if !strings.Contains(source, marker) {
					t.Fatalf("%s lacks %q", test.name, marker)
				}
			}
			for _, forbidden := range []string{
				"DISABLE TRIGGER", "session_replication_role",
				"GRANT INSERT", "TRUNCATE ", "-- +goose Down",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains %q", test.name, forbidden)
				}
			}
		})
	}
}

func TestBillingInstallationAssignmentUsesLockedCurrentAuthority(t *testing.T) {
	const name = "20260729100007_harden_billing_installation_assign_v2"
	source := pairedBillingAuthorityMigration(t, name)
	for _, marker := range []string{
		"DROP FUNCTION platform.billing_installation_assign_v2(",
		"INTO STRICT v_group_generation",
		"FOR SHARE OF tenant, group_row, billing_member, forms_member",
		"billing_assert_tenant_authority_v1(",
		"billing installation base assignment is unavailable",
		"p_authority_id, p_tenant_id, v_group_generation",
		"p_binding_purpose = 'onpremise_primary'",
		"p_binding_purpose IN ('saas_control','saas_customer')",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("installation authority hardening lacks %q", marker)
		}
	}
	if strings.Contains(source, "p_authority_id,p_tenant_id,1,1,1") {
		t.Fatal("installation authority still hard-codes generation one")
	}
}

func TestBillingInstallationAssignmentRequiresTenancyCapability(t *testing.T) {
	const name = "20260729100016_require_billing_installation_capability_v1"
	source := pairedBillingAuthorityMigration(t, name)
	for _, marker := range []string{
		"CREATE FUNCTION platform.billing_installation_assign_v3(",
		"billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text",
		"billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text",
		"TO harpy_runtime",
		"billing_installation_assignment_contexts",
		"billing_installation_id_v1(",
		"'tenant-authority'",
		"'base-offering-assignment'",
		"billing.assign_tenant_base_offering",
		"tenancy_activation_command_consume_v1(",
		") IS DISTINCT FROM TRUE THEN",
		"billing installation assignment profile is invalid",
		"billing installation assignment replay differs",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("installation capability hardening lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"billing_installation_assign_v4", "DISABLE TRIGGER",
		"session_replication_role", "GRANT INSERT",
		"billing_installation_base_assignments_context_fk",
		"DEFERRABLE INITIALLY DEFERRED",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("installation capability hardening contains %q", forbidden)
		}
	}
}

func TestBillingReconciliationRequiresOwnedCapabilityScope(t *testing.T) {
	tests := map[string]string{
		"20260729100019_require_billing_global_reconcile_scope_v1": "billing_reconcile_current_global_authority_v2",
		"20260729100020_require_billing_tenant_reconcile_scope_v1": "billing_reconcile_current_tenant_authorities_v2",
	}
	for name, successor := range tests {
		source := pairedBillingAuthorityMigration(t, name)
		for _, marker := range []string{
			"CREATE FUNCTION platform." + successor + "()",
			"tenancy_authority_reconcile_scope_assert_v1(",
			"'billing'", ") IS DISTINCT FROM TRUE THEN",
			"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
			"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		} {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s lacks %q", name, marker)
			}
		}
		for _, forbidden := range []string{
			"DROP FUNCTION platform.",
			"TO harpy_runtime",
			"REVOKE EXECUTE ON FUNCTION",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s violates additive expand with %q", name, forbidden)
			}
		}
	}
}

func TestBillingFinalizedSuccessorIsNarrowAndEvidenceBound(t *testing.T) {
	const name = "20260729100008_create_billing_finalized_authority_successor_v1"
	source := pairedBillingAuthorityMigration(t, name)
	for _, marker := range []string{
		"billing_finalize_authority_successor_v1",
		"SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer'",
		"operation.state = 'completed'",
		"operation.domain = 'billing'",
		"operation.domain = 'forms'",
		"operation.cutover_generation = v_operation.cutover_generation",
		"group_row.generation = v_operation.cutover_generation + 1",
		"group_row.acceptance_digest = p_acceptance_digest",
		"domain_row.acceptance_digest = p_acceptance_digest",
		"group_row.approved_by IS NOT NULL",
		"group_row.approved_by <> group_row.requested_by",
		"domain_row.requested_by = v_requested_by",
		"domain_row.approved_by = v_approved_by",
		"domain_row.reason = v_reason",
		"domain_row.change_reference = v_change_reference",
		"domain_row.changed_at = v_changed_at",
		"pg_catalog.uuid_extract_version(p_operation_id)",
		"v_run_hash := pg_catalog.encode(pg_catalog.sha256(",
		"authority.authority_id = v_authority_id",
		"EXCEPTION WHEN SQLSTATE '23505'",
		"TO harpy_cutover_authority_owner",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("finalized Billing successor lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"GRANT INSERT", "TO harpy_cutover_finalizer",
		"DISABLE TRIGGER", "session_replication_role",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("finalized Billing successor contains %q", forbidden)
		}
	}
}

func pairedBillingAuthorityMigration(t *testing.T, name string) string {
	t.Helper()
	fresh := readBillingSnapshotMigration(t, "fresh", name+".fresh.sql")
	adopted := readBillingSnapshotMigration(
		t, "v1_adopted", name+".v1_adopted.sql",
	)
	if !bytes.Equal(fresh, adopted) {
		t.Fatalf("%s migration variants differ", name)
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("%s migration is invalid: %v", name, err)
	}
	return string(fresh)
}
