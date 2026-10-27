package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogDemoReleaseFrontierIsPairedAndFailClosed(t *testing.T) {
	const base = "20260730144010_assert_catalog_demo_release_frontier_v1"
	fresh, err := os.ReadFile(filepath.Join(
		"migrations", "platform", "fresh", base+".fresh.sql",
	))
	if err != nil || len(fresh) == 0 {
		t.Fatalf("Catalog demo release frontier did not load: %v", err)
	}
	document := string(fresh)
	for _, required := range []string{
		"SELECT 1 / CASE WHEN",
		"BETWEEN 180000 AND 189999",
		"'platform.catalog_package_owner_payloads'",
		"role_record.rolname = 'harpy_runtime'",
		"role_record.rolname = 'harpy_v2_runtime'",
		"NOT role_record.rolsuper",
		"NOT role_record.rolbypassrls",
		"THEN 1 ELSE 0 END",
	} {
		if !strings.Contains(document, required) {
			t.Fatalf("Catalog demo release frontier lacks %q", required)
		}
	}
	if strings.Contains(document, "CREATE TABLE") ||
		strings.Contains(document, "CREATE ROLE") {
		t.Fatal("Catalog demo release frontier repairs a missing prerequisite")
	}
}
