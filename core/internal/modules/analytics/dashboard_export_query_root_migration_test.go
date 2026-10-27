package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDashboardExportQueryRootMigrationIsPairedAndFailClosed(t *testing.T) {
	const logical = "20260802010037_resolve_analytics_query_root_by_version_v1"
	fresh, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "fresh", logical+".fresh.sql",
	))
	if err != nil || len(fresh) == 0 {
		t.Fatalf("dashboard export query-root migration did not load: %v", err)
	}
	text := string(fresh)
	for _, required := range []string{
		"SECURITY DEFINER SET search_path FROM CURRENT",
		"analytics_assert_current_authority_v1(p_scope_schema)",
		"version_row.version_state='published'",
		"root.lifecycle_state='active'",
		"REVOKE ALL ON FUNCTION analytics_query_root_by_version_v1",
		"TO harpy_runtime",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("query-root migration is missing %q", required)
		}
	}
}
