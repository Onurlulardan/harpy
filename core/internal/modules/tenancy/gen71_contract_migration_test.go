package tenancy

import (
	"strings"
	"testing"
)

func TestGen71TenancyContractRetiresOnlyLegacyWriters(t *testing.T) {
	source := pairedTenancyMigration(
		t, "20260729100037_contract_gen71_tenancy_legacy_writers_v1",
	)
	retired := []string{
		"platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)",
		"platform.tenancy_onboarding_activate_v1(uuid,text,bigint)",
		"platform.tenancy_onboarding_activate_v2(uuid,text,bigint)",
		"platform.tenancy_activation_initialize_v1(uuid,uuid,uuid,text,text,uuid,text,text,text)",
		"platform.tenancy_activation_initialize_v2(uuid,uuid,uuid,text,text,uuid,text,text,text)",
		"platform.tenancy_activation_initialize_v3(uuid,uuid,uuid,text,text,uuid,text,text,text)",
		"platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text)",
		"platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint)",
		"platform.tenancy_activation_assert_v1(uuid,uuid,name)",
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
				t.Fatalf("gen71 tenancy contract lacks %q", marker)
			}
		}
		if strings.Contains(source, "DROP FUNCTION "+signature) {
			t.Fatalf("gen71 tenancy contract drops %s", signature)
		}
	}
	for _, retained := range []string{
		"platform.tenancy_onboarding_accept_v2(",
		"platform.tenancy_onboarding_activate_v3(",
		"platform.tenancy_activation_initialize_v5(",
		"platform.tenancy_onboarding_complete_v2(",
		"platform.tenancy_activation_assert_v2(",
	} {
		if !strings.Contains(source, retained) {
			t.Fatalf("gen71 tenancy contract lacks successor %s", retained)
		}
		if strings.Contains(source, "REVOKE EXECUTE ON FUNCTION "+retained) {
			t.Fatalf("gen71 tenancy contract revokes successor %s", retained)
		}
	}
	for _, marker := range []string{
		"pg_catalog.has_function_privilege(",
		"'harpy_v2_runtime', routine_oid, 'EXECUTE'",
		"platform.tenancy_activation_command_capabilities",
		"platform.tenancy_authority_reconcile_scopes",
		"platform.tenancy_activation_step_capabilities",
		"THEN 0 ELSE 1 END",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("gen71 tenancy fence lacks %q", marker)
		}
	}
}
