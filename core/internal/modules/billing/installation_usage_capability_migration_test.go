package billing

import (
	"strings"
	"testing"
)

func TestBillingInstallationUsageOpeningUsesAdditiveCapabilitySuccessor(
	t *testing.T,
) {
	const name = "20260729100026_create_billing_installation_open_usage_v2"
	source := pairedBillingAuthorityMigration(t, name)
	for _, marker := range []string{
		"CREATE FUNCTION platform.billing_installation_open_usage_v2(",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"platform.tenancy_activation_step_consume_v1(",
		"'billing.open_tenant_usage'",
		"platform.billing_installation_open_usage_v1(",
		"TO harpy_runtime",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Billing usage capability successor lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"DROP FUNCTION platform.billing_installation_open_usage_v1",
		"CREATE OR REPLACE FUNCTION platform.billing_installation_open_usage_v1",
		"REVOKE EXECUTE ON FUNCTION platform.billing_installation_open_usage_v1",
		"DISABLE TRIGGER",
		"session_replication_role",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("Billing usage expand contains %q", forbidden)
		}
	}
}
