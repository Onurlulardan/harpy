package analytics

import (
	"bytes"
	"strings"
	"testing"
)

func TestDashboardExportQueryRootMigrationIsPairedAndFailClosed(t *testing.T) {
	const logical = "20260802010037_resolve_analytics_query_root_by_version_v1"
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	contents := map[string][]byte{}
	for _, document := range documents {
		if document.Scope() != "tenant" || !strings.Contains(document.Path(), logical) {
			continue
		}
		origin, _ := analyticsMigrationIdentity(t, document.Scope(), document.Path())
		contents[origin] = document.Content()
	}
	if len(contents) != 2 || !bytes.Equal(contents["fresh"], contents["v1_adopted"]) {
		t.Fatal("dashboard export query-root migration pair differs")
	}
	text := string(contents["fresh"])
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
