package tenancy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestTenancyTransactionCapabilitiesLoadFreshAndFailClosed(t *testing.T) {
	capability := freshTenancyMigration(
		t, "20260729100013_create_tenancy_transaction_capabilities_v1",
	)
	for _, marker := range []string{
		"platform.tenancy_activation_command_capabilities",
		"platform.tenancy_authority_reconcile_scopes",
		"pg_catalog.pg_backend_pid()",
		"pg_catalog.pg_current_xact_id()",
		"platform.tenancy_activation_command_consume_v1(",
		"platform.tenancy_authority_reconcile_scope_assert_v1(",
		"DEFERRABLE INITIALLY DEFERRED",
		"was not consumed", "was not cleared",
	} {
		if !strings.Contains(capability, marker) {
			t.Fatalf("tenancy capability migration lacks %q", marker)
		}
	}
	activation := freshTenancyMigration(
		t, "20260729100014_create_tenancy_onboarding_activate_v2",
	)
	for _, marker := range []string{
		"platform.tenancy_onboarding_activate_v2(",
		"platform.tenancy_onboarding_activate_v1(",
		"INSERT INTO platform.tenancy_activation_command_capabilities",
		"'billing.assign_tenant_base_offering'",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v2(",
	} {
		if !strings.Contains(activation, marker) {
			t.Fatalf("tenancy activation v2 migration lacks %q", marker)
		}
	}
	if strings.Contains(activation,
		"REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v1(") {
		t.Fatal("gen70 activation migration revokes the gen23 ABI")
	}
	reconcile := freshTenancyMigration(
		t, "20260729100015_create_tenancy_authority_reconcile_v1",
	)
	for _, marker := range []string{
		"platform.tenancy_authority_reconcile_current_v1(",
		"v_operation.state IS DISTINCT FROM 'succeeded'",
		"FROM platform.billing_reconcile_current_global_authority_v2()",
		"FROM platform.billing_reconcile_current_tenant_authorities_v2()",
		"FROM platform.workflow_reconcile_current_tenant_authorities_v2()",
		"FROM platform.analytics_reconcile_current_tenant_authorities_v2()",
		"DELETE FROM platform.tenancy_authority_reconcile_scopes",
		"v_deleted IS DISTINCT FROM 3",
	} {
		if !strings.Contains(reconcile, marker) {
			t.Fatalf("tenancy reconcile wrapper lacks %q", marker)
		}
	}
}

func freshTenancyMigration(t *testing.T, name string) string {
	t.Helper()
	fresh, err := os.ReadFile(filepath.Join(
		"migrations", "platform", "fresh", name+".fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh) == 0 {
		t.Fatalf("%s fresh migration is empty", name)
	}
	if _, err = postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("%s migration is invalid: %v", name, err)
	}
	return string(fresh)
}
