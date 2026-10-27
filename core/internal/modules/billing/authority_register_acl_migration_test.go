package billing

import (
	"strings"
	"testing"
)

func TestBillingAuthorityExpandPreservesGeneration23RoutineACLs(t *testing.T) {
	tests := map[string][]string{
		"20260729100016_require_billing_installation_capability_v1": {
			"platform.billing_installation_assign_v1(",
			"platform.billing_installation_assign_v2(",
		},
		"20260729100019_require_billing_global_reconcile_scope_v1": {
			"platform.billing_register_global_authority_v1(",
		},
		"20260729100020_require_billing_tenant_reconcile_scope_v1": {
			"platform.billing_register_tenant_authority_v1(",
		},
	}
	for migration, routines := range tests {
		source := pairedBillingAuthorityMigration(t, migration)
		for _, routine := range routines {
			if !strings.Contains(source, routine) {
				t.Fatalf("%s does not reference legacy runtime %s",
					migration, routine)
			}
			if strings.Contains(source, "REVOKE EXECUTE ON FUNCTION "+routine) ||
				strings.Contains(source, "REVOKE EXECUTE ON FUNCTION\n    "+routine) {
				t.Fatalf("%s revokes generation-23 runtime %s",
					migration, routine)
			}
		}
	}
}
