package analytics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const dashboardDefaultReadMigration = "20260728110003_" +
	"create_analytics_dashboard_default_read_v1"

func TestDashboardDefaultReadMigrationIsHardenedAndPaired(t *testing.T) {
	fresh, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "fresh", dashboardDefaultReadMigration+".fresh.sql",
	))
	if err != nil || len(fresh) == 0 {
		t.Fatalf("default dashboard owner-read migration did not load: %v", err)
	}
	source := string(fresh)
	for _, required := range []string{
		"SECURITY DEFINER",
		"SET search_path FROM CURRENT",
		"analytics_assert_current_authority_v1(p_scope_schema)",
		"REVOKE ALL ON FUNCTION analytics_dashboard_get_default_v1",
		"GRANT EXECUTE ON FUNCTION analytics_dashboard_get_default_v1",
		"TO harpy_runtime",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("default dashboard owner-read migration lacks %q", required)
		}
	}
	if strings.Contains(source, "GRANT SELECT") {
		t.Fatal("default dashboard owner-read migration grants direct table access")
	}
}
