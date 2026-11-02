package async

import (
	"os"
	"strings"
	"testing"
)

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
