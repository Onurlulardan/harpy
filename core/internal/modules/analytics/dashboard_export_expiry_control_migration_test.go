package analytics

import (
	"bytes"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestDashboardExportExpiryControlQualifiesStateOutputCollision(t *testing.T) {
	const logical = "20260802010028_create_analytics_dashboard_export_expiry_control_v1"
	fresh := readDashboardExportPlatformMigration(t, "fresh", logical)
	if len(fresh) == 0 {
		t.Fatal("dashboard export expiry control migration is empty")
	}
	if bytes.Count(fresh, []byte{'\n'}) > 200 {
		t.Fatal("dashboard export expiry control migration exceeds 200 lines")
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("dashboard export expiry control migration is invalid: %v", err)
	}
	normalized := strings.Join(strings.Fields(string(fresh)), " ")
	if !strings.Contains(normalized,
		"ELSE analytics_dashboard_exports.state END") ||
		strings.Contains(normalized, "ELSE state END") {
		t.Fatal("dashboard export expiry control state RHS is output-variable ambiguous")
	}
}
