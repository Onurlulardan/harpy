package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestAnalyticsCutoverSuccessorAuthorityIsPairedAndFailClosed(t *testing.T) {
	const name = "20260729100009_create_analytics_cutover_successor_authority_v1"
	fresh := readAnalyticsCutoverSuccessor(t, "fresh", name+".fresh.sql")
	if len(fresh) == 0 {
		t.Fatal("analytics cutover successor authority migration is empty")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("analytics cutover successor authority migration is invalid: %v", err)
	}
	source := string(fresh)
	for _, marker := range []string{
		"CREATE FUNCTION platform.analytics_guard_tenant_authority_v2()",
		"SESSION_USER IS NOT DISTINCT FROM 'harpy_cutover_finalizer'",
		"CREATE OR REPLACE TRIGGER analytics_tenant_authorities_guard",
		"ENABLE ALWAYS TRIGGER analytics_tenant_authorities_guard",
		"CREATE FUNCTION platform.analytics_finalize_authority_successor_v1(",
		"SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer'",
		"operation.operation_id = p_operation_id AND operation.run_id = p_run_id",
		"operation.state = 'completed'",
		"operation.group_id = 'analytics' AND operation.domain = 'analytics'",
		"group_row.generation = operation.cutover_generation + 1",
		"domain_row.generation = operation.cutover_generation + 1",
		"group_row.acceptance_digest = p_acceptance_digest",
		"domain_row.acceptance_digest = p_acceptance_digest",
		"extra.group_id = operation.group_id",
		"pg_catalog.sha256(",
		"ON CONFLICT DO NOTHING",
		"TO harpy_cutover_authority_owner",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("analytics cutover successor authority lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"CREATE OR REPLACE FUNCTION", "DISABLE TRIGGER", "session_replication_role",
		"GRANT INSERT", "TO harpy_cutover_finalizer",
		"cutover_authority.finalization_receipts", "verify_finalization_receipt_v1",
		"TRUNCATE ", "-- +goose Down",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("analytics cutover successor authority contains %q", forbidden)
		}
	}
}

func readAnalyticsCutoverSuccessor(t *testing.T, origin, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(
		"migrations", "platform", origin, name,
	))
	if err != nil {
		t.Fatal(err)
	}
	return content
}
