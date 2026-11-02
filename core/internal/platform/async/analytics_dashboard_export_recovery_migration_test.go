package async

import (
	"os"
	"strings"
	"testing"
)

func analyticsExportFresh(t *testing.T, logical string) string {
	t.Helper()
	document := analyticsDashboardExportMigrationPairs(t)[logical]["fresh"]
	if len(document) == 0 {
		t.Fatalf("missing Analytics dashboard export migration %s", logical)
	}
	return string(document)
}

func TestGenericExpiredRecoveryExcludesAnalyticsDashboardExport(t *testing.T) {
	document, err := os.ReadFile("adapters/postgres/sql/lock_expired_jobs.sql")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(document),
		"job.kind <> 'analytics.dashboard.export.render'") {
		t.Fatal("generic expired recovery can consume Analytics dashboard export jobs")
	}
}
