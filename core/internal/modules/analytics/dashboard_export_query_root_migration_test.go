package analytics

import (
	"bytes"
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
	if err != nil {
		t.Fatal(err)
	}
	adopted, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "v1_adopted", logical+".v1_adopted.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fresh, adopted) {
		t.Fatal("dashboard export query-root migration pair differs")
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
