package billing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestBillingCurrentAuthorityResolverIsSnapshotSafe(t *testing.T) {
	const name = "20260728110000_make_billing_current_authority_resolver_snapshot_safe_v1"
	fresh := readBillingSnapshotMigration(t, "fresh", name+".fresh.sql")
	if len(fresh) == 0 {
		t.Fatal("billing snapshot-safe resolver migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("billing snapshot-safe resolver migration is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"DROP FUNCTION platform.billing_resolve_current_tenant_authority_v1(uuid)",
		"CREATE FUNCTION platform.billing_resolve_current_tenant_authority_v1",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"CURRENT_USER IS DISTINCT FROM 'harpy_owner'",
		"tenant.status IN ('ACTIVE', 'TRIAL')",
		"authority.authority_group_id = 'billing'",
		"billing_member.domain = 'billing' AND billing_member.ordinal = 1",
		"group_row.state = 'v2' AND group_row.writer = 'v2'",
		"billing_row.state = 'v2' AND billing_row.writer = 'v2'",
		"authority.authority_group_generation = group_row.generation",
		"authority.billing_generation = billing_row.generation",
		"NOT EXISTS (",
		"REVOKE ALL ON FUNCTION",
		"GRANT EXECUTE ON FUNCTION",
		"TO harpy_runtime;",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("billing snapshot-safe resolver migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"FOR SHARE", "FOR UPDATE", "billing_assert_tenant_authority_v1",
		"INSERT ", "UPDATE ", "DELETE ", "-- +goose Down", "CREATE OR REPLACE",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("billing snapshot-safe resolver contains %q", forbidden)
		}
	}
}

func readBillingSnapshotMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(
		"migrations", "platform", origin, name,
	))
	if err != nil {
		t.Fatal(err)
	}
	return content
}
