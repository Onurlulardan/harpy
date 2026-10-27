package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestAnalyticsCurrentAuthorityReconcileIsPairedAndFailClosed(t *testing.T) {
	const name = "20260729100000_create_analytics_authority_reconcile_v1"
	fresh := readAnalyticsAuthorityReconcile(t, "fresh", name+".fresh.sql")
	if len(fresh) == 0 {
		t.Fatal("analytics current-authority reconcile migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("analytics current-authority reconcile is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"DROP FUNCTION platform.analytics_resolve_current_tenant_authority_v1(uuid);",
		"CREATE FUNCTION platform.analytics_resolve_current_tenant_authority_v1(",
		"INTO STRICT v_authority_id, v_group_generation, v_domain_generation",
		"EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN",
		"CREATE FUNCTION platform.analytics_reconcile_current_tenant_authorities_v1()",
		"SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'",
		"'harpy-v2:core-operator-initialize'",
		"pg_catalog.pg_advisory_xact_lock",
		"platform.analytics_register_tenant_authority_v1(",
		"pg_catalog.uuidv7()",
		"group_row.state = 'v2' AND group_row.writer = 'v2'",
		"domain_row.state = 'v2' AND domain_row.writer = 'v2'",
		"authority_row.authority_mode = 'runtime_v2'",
		"EXCEPTION WHEN SQLSTATE 'P6101' THEN",
		"extra.group_id = 'analytics'",
		"analytics authority reconciliation is incomplete",
		"REVOKE ALL ON FUNCTION",
		"TO harpy_runtime",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("analytics current-authority reconcile lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"DISABLE TRIGGER", "session_replication_role", "CREATE OR REPLACE",
		"INSERT INTO platform.analytics_tenant_authorities", "DROP TABLE",
		"DROP TRIGGER", "TRUNCATE ", "-- +goose Down",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("analytics current-authority reconcile contains %q", forbidden)
		}
	}
}

func readAnalyticsAuthorityReconcile(t *testing.T, origin, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(
		"migrations", "platform", origin, name,
	))
	if err != nil {
		t.Fatal(err)
	}
	return content
}
