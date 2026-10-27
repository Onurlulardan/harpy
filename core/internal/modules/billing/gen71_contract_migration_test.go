package billing

import (
	"strings"
	"testing"
)

func TestGen71BillingContractRetiresOnlyLegacyWriters(t *testing.T) {
	source := pairedBillingAuthorityMigration(
		t, "20260729100039_contract_gen71_billing_legacy_writers_v1",
	)
	retired := []string{
		"platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text)",
		"platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)",
		"platform.billing_installation_open_usage_v1(uuid,uuid,text)",
		"platform.billing_reconcile_current_global_authority_v1()",
		"platform.billing_reconcile_current_tenant_authorities_v1()",
		"platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz)",
		"platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,bigint,timestamptz)",
	}
	for _, signature := range retired {
		for _, marker := range []string{
			"REVOKE ALL ON FUNCTION " + signature + " FROM PUBLIC",
			"REVOKE EXECUTE ON FUNCTION " + signature +
				" FROM harpy_runtime",
			"REVOKE EXECUTE ON FUNCTION " + signature +
				" FROM harpy_v2_runtime",
		} {
			if !strings.Contains(source, marker) {
				t.Fatalf("gen71 Billing contract lacks %q", marker)
			}
		}
		if strings.Contains(source, "DROP FUNCTION "+signature) {
			t.Fatalf("gen71 Billing contract drops %s", signature)
		}
	}
	for _, retained := range []string{
		"platform.billing_installation_assign_v3(",
		"platform.billing_installation_open_usage_v2(",
	} {
		if !strings.Contains(source, retained) {
			t.Fatalf("gen71 Billing contract lacks successor %s", retained)
		}
		if strings.Contains(source, "REVOKE EXECUTE ON FUNCTION "+retained) {
			t.Fatalf("gen71 Billing contract revokes successor %s", retained)
		}
	}
	for _, marker := range []string{
		"pg_catalog.has_function_privilege(",
		"'harpy_v2_runtime', routine_oid, 'EXECUTE'",
		"platform.billing_installation_id_v1(uuid,text)",
		"platform.billing_reconcile_current_global_authority_v2()",
		"platform.billing_reconcile_current_tenant_authorities_v2()",
		"platform.billing_installation_assignment_contexts",
		"THEN 0 ELSE 1 END",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("gen71 Billing fence lacks %q", marker)
		}
	}
}
