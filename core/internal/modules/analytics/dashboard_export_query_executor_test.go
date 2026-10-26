package analytics

import (
	"testing"
)

func TestDashboardExportParametersMergeDeterministically(t *testing.T) {
	merged, err := mergeDashboardExportParameters(
		[]byte(`{"region":"eu","limit":10}`),
		[]byte(`{"limit":20,"status":"open"}`),
	)
	if err != nil || string(merged) != `{"limit":20,"region":"eu","status":"open"}` {
		t.Fatalf("merged parameters = %s, %v", merged, err)
	}
}

func TestDashboardExportResultRowsRequiresCanonicalQueryEnvelope(t *testing.T) {
	rows, err := dashboardExportResultRows([]byte(
		`{"items":[{"id":1},{"id":2}],"pagination":{"size":2}}`,
	))
	if err != nil || rows != 2 {
		t.Fatalf("rows = %d, %v", rows, err)
	}
	if _, err = dashboardExportResultRows([]byte(`{"data":[]}`)); err == nil {
		t.Fatal("non-query result envelope was accepted")
	}
}
