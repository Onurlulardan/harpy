package analytics

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestAnalyticsCurrentAuthorityResolverIsSnapshotSafe(t *testing.T) {
	const name = "20260728110002_make_analytics_current_authority_resolver_snapshot_safe_v1"
	fresh := readAnalyticsSnapshotMigration(t, "fresh", name+".fresh.sql")
	adopted := readAnalyticsSnapshotMigration(t, "v1_adopted",
		name+".v1_adopted.sql")
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("analytics snapshot-safe resolver migration variants differ")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("analytics snapshot-safe resolver migration is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"DROP FUNCTION platform.analytics_resolve_current_tenant_authority_v1(uuid)",
		"CREATE FUNCTION platform.analytics_resolve_current_tenant_authority_v1",
		"SECURITY DEFINER", "SET search_path = pg_catalog",
		"v_mode := 'runtime_v2'; v_state := 'v2'; v_writer := 'v2'",
		"v_mode := 'adopter_shadow'; v_state := 'shadow'; v_writer := 'v1'",
		"tenant.status IN ('ACTIVE', 'TRIAL')",
		"authority_row.authority_group_id = 'analytics'",
		"authority_row.authority_domain = 'analytics'",
		"authority_row.authority_mode = v_mode",
		"authority_row.authority_group_generation = group_row.generation",
		"authority_row.authority_domain_generation = domain_row.generation",
		"group_row.state = v_state AND group_row.writer = v_writer",
		"domain_row.state = v_state AND domain_row.writer = v_writer",
		"ERRCODE = 'P6101'", "REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("analytics snapshot-safe resolver migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"FOR SHARE", "FOR UPDATE", "INSERT ", "UPDATE ", "DELETE ",
		"-- +goose Down", "GRANT EXECUTE", "CREATE OR REPLACE",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("analytics snapshot-safe resolver contains %q", forbidden)
		}
	}
}

func readAnalyticsSnapshotMigration(t *testing.T, origin, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(
		"migrations", "platform", origin, name,
	))
	if err != nil {
		t.Fatal(err)
	}
	return content
}
