package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	emptyMenuBaselineVersion = "20260728103100"
	emptyMenuRepairVersion   = "20260728105100"
)

func TestEmptyMenuBaselineMigrationIsPairedAndDataPreserving(t *testing.T) {
	source := catalogMenuMigration(t, emptyMenuBaselineVersion)
	for _, marker := range []string{
		"ON CONFLICT (definition_id) DO NOTHING",
		"definition.active_version_id IS NULL",
		"NOT EXISTS (",
		"ON CONFLICT (version_id) DO NOTHING",
		"UPDATE catalog_menu_definitions",
		"AND active_version_id IS NULL",
		"JOIN catalog_menu_versions AS version",
		"FROM catalog_menu_items AS item",
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Catalog empty menu baseline migration lacks %q", marker)
		}
	}
	if strings.Count(source, "-- +goose Up") != 1 ||
		strings.Contains(source, "-- +goose Down") ||
		strings.Contains(source, "DELETE FROM") {
		t.Fatal("Catalog empty menu baseline migration is not forward-only")
	}
}

func TestEmptyMenuV1TargetRepairIsPairedAndFailClosed(t *testing.T) {
	source := catalogMenuMigration(t, emptyMenuRepairVersion)
	for _, marker := range []string{
		"CREATE FUNCTION catalog_retract_empty_menu_from_v1_target_v1(",
		"p_expected_owner IS DISTINCT FROM 'harpy_owner'::name",
		"IF v_origin = 'v1_adopted' THEN",
		"FROM platform.cutover_groups AS authority",
		"FOR SHARE",
		"FROM platform.cutover_group_members AS member",
		"FROM platform.domain_cutovers AS domain",
		"v_domain_generation IS DISTINCT FROM v_group_generation",
		"IF v_group_state = 'v2' AND v_group_writer = 'v2' THEN",
		"v_group_state IN ('v1', 'shadow')",
		"v_group_state IN ('quiescing', 'fenced')",
		"IF v_existing <> 2 OR NOT EXISTS",
		"definition.revision = 1",
		"version.version_number = 1",
		"FROM catalog_menu_items AS item",
		"Catalog V1 inactive target is not the exact baseline",
		"SET active_version_id = NULL",
		"DELETE FROM catalog_menu_versions",
		"DELETE FROM catalog_menu_definitions",
		"DROP FUNCTION catalog_retract_empty_menu_from_v1_target_v1(name)",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Catalog empty menu target repair lacks %q", marker)
		}
	}
	if strings.Count(source, "-- +goose Up") != 1 ||
		strings.Contains(source, "-- +goose Down") {
		t.Fatal("Catalog empty menu target repair is not forward-only")
	}
}

func catalogMenuMigration(t *testing.T, version string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(
		"migrations", "tenant", "fresh", version+"_*.fresh.sql",
	))
	if err != nil || len(matches) != 1 {
		t.Fatalf("catalog menu migration %s fresh count = %d err=%v",
			version, len(matches), err)
	}
	fresh, err := os.ReadFile(matches[0])
	if err != nil || len(fresh) == 0 {
		t.Fatalf("catalog menu migration %s did not load: %v", version, err)
	}
	return string(fresh)
}
