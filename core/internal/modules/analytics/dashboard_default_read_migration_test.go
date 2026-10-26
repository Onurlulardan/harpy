package analytics

import (
	"bytes"
	"strings"
	"testing"
)

const dashboardDefaultReadMigration = "20260728110003_" +
	"create_analytics_dashboard_default_read_v1"

func TestDashboardDefaultReadMigrationIsHardenedAndPaired(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var fresh, adopted []byte
	for _, document := range documents {
		if !strings.Contains(document.Path(), dashboardDefaultReadMigration) {
			continue
		}
		if strings.Contains(document.Path(), "/fresh/") {
			fresh = document.Content()
		} else if strings.Contains(document.Path(), "/v1_adopted/") {
			adopted = document.Content()
		}
	}
	if len(fresh) == 0 || !bytes.Equal(fresh, adopted) {
		t.Fatal("default dashboard owner-read migration pair is missing or differs")
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
