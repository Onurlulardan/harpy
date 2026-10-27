package analytics

import (
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestAnalyticsAuthorityReconcileV2IsPairedAndCapabilityScoped(t *testing.T) {
	const name = "20260729100018_create_analytics_authority_reconcile_v2"
	fresh := readAnalyticsAuthorityReconcile(t, "fresh", name+".fresh.sql")
	if len(fresh) == 0 {
		t.Fatal("analytics authority reconcile v2 migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("analytics authority reconcile v2 is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.analytics_reconcile_current_tenant_authorities_v2()",
		"platform.tenancy_authority_reconcile_scope_assert_v1('analytics')",
		"IS DISTINCT FROM true",
		"platform.analytics_register_tenant_authority_v1(",
		"EXCEPTION WHEN SQLSTATE 'P6101'",
		"REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("analytics authority reconcile v2 lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"DROP FUNCTION", "CREATE OR REPLACE", "TO harpy_runtime",
		"REVOKE EXECUTE ON FUNCTION platform.analytics_register",
		"-- +goose Down",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("analytics authority reconcile v2 contains %q", forbidden)
		}
	}
}
