package async

import (
	"path"
	"strconv"
	"strings"
	"testing"
)

const analyticsDashboardExportMigrationStart = int64(20260802020000)
const analyticsDashboardExportMigrationEnd = int64(20260802020049)

func analyticsDashboardExportFreshMigrations(t *testing.T) map[string][]byte {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	fresh := map[string][]byte{}
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasSuffix(name, ".fresh.sql") ||
			strings.Contains(document.Path(), "/supersessions/") {
			continue
		}
		version, parseErr := strconv.ParseInt(name[:14], 10, 64)
		if parseErr != nil || version < analyticsDashboardExportMigrationStart ||
			version > analyticsDashboardExportMigrationEnd {
			continue
		}
		fresh[strings.TrimSuffix(name, ".fresh.sql")] = document.Content()
	}
	return fresh
}

func TestAnalyticsDashboardExportPlatformMigrationsLoadFresh(t *testing.T) {
	fresh := analyticsDashboardExportFreshMigrations(t)
	if len(fresh) != 50 {
		t.Fatalf("Analytics dashboard export fresh migrations=%d, want 50", len(fresh))
	}
	for logical, document := range fresh {
		if len(document) == 0 {
			t.Fatalf("fresh analytics dashboard export migration did not load: %s", logical)
		}
	}
}

func TestAnalyticsDashboardExportClaimAndTerminalContracts(t *testing.T) {
	fresh := analyticsDashboardExportFreshMigrations(t)
	claim := string(fresh["20260802020001_create_analytics_dashboard_export_claim_v1"])
	for _, token := range []string{"analytics.dashboard.export.render", "quota_wait",
		"2147483648", "21474836480", "uuid_extract_version(p_lease_owner_id)",
		"INSERT INTO platform.job_attempts"} {
		if !strings.Contains(claim, token) {
			t.Fatalf("claim migration misses %q", token)
		}
	}
	if strings.Index(claim, "claim_state:='quota_wait'") >
		strings.Index(claim, "UPDATE platform.artifact_quota_counters") {
		t.Fatal("quota wait occurs after claim mutation")
	}
	finish := string(fresh["20260802020005_create_analytics_dashboard_export_finish_v1"])
	for _, token := range []string{"RETURNS TABLE(job_state text,effect_receipt_id uuid)",
		"v_attempt.effect_started_at IS NULL", "cancellation cleanup is required",
		"v_attempt.finished_at IS NOT NULL", "effect_receipt_id:=v_attempt.effect_receipt_id"} {
		if !strings.Contains(finish, token) {
			t.Fatalf("finish migration misses %q", token)
		}
	}
	if strings.Index(finish, "RETURN NEXT; RETURN;") >
		strings.Index(finish, "UPDATE platform.artifact_quota_counters") {
		t.Fatal("finish replay reaches quota mutation")
	}
	succeed := string(fresh["20260802020004_create_analytics_dashboard_export_succeed_v1"])
	if strings.Index(succeed, "RETURN true;") >
		strings.Index(succeed, "UPDATE platform.artifact_quota_counters") {
		t.Fatal("success replay reaches quota mutation")
	}
}

func TestAnalyticsDashboardExportUnknownRecoveryContract(t *testing.T) {
	fresh := analyticsDashboardExportFreshMigrations(t)
	unknown := string(fresh["20260802020006_create_analytics_dashboard_export_unknown_v1"])
	if !strings.Contains(unknown, "JOB_OUTCOME_UNKNOWN") ||
		strings.Contains(unknown, "artifact_quota_counters") {
		t.Fatal("unknown terminal must retain quota")
	}
	resolve := string(fresh["20260802020024_create_analytics_dashboard_export_unknown_resolve_v2"])
	for _, token := range []string{"resolution_kind NOT IN ('deleted','absent')",
		"analytics_dashboard_export_unknown_resolutions", "JOB_OUTCOME_UNKNOWN",
		"unknown_cleanup_claims", "p_cleanup_generation",
		"reserved_ciphertext_bytes-2147483648"} {
		if !strings.Contains(resolve, token) {
			t.Fatalf("unknown resolution migration misses %q", token)
		}
	}
	if strings.Contains(resolve, "UPDATE platform.jobs") ||
		strings.Contains(resolve, "UPDATE platform.job_attempts") {
		t.Fatal("unknown resolution reverses immutable terminal evidence")
	}
	owner := string(fresh["20260802020014_create_analytics_dashboard_export_cleanup_owner_v1"])
	if !strings.Contains(owner, "cleanup owner is not current") ||
		strings.Contains(owner, "authority_group_generation") ||
		strings.Contains(owner, "cutover_generation") {
		t.Fatal("cleanup owner is not current-writer fenced independently of frozen generations")
	}
	candidate := string(fresh["20260802020011_create_analytics_dashboard_export_unknown_candidate_v1"])
	for _, token := range []string{"group_row.state='v2'", "domain_row.state='v2'",
		"cutover_group_members", "NOT EXISTS", "ORDER BY job.completed_at,job.id LIMIT 1"} {
		if !strings.Contains(candidate, token) {
			t.Fatalf("unknown candidate migration misses %q", token)
		}
	}
	for _, forbidden := range []string{
		"group_row.generation=job.authority_group_generation",
		"domain_row.generation=job.cutover_generation",
	} {
		if strings.Contains(candidate, forbidden) {
			t.Fatalf("stale generation cleanup is blocked by %q", forbidden)
		}
	}
}

func TestAnalyticsDashboardExportCrashBoundaryContract(t *testing.T) {
	fresh := analyticsDashboardExportFreshMigrations(t)
	resolve := string(fresh["20260802020024_create_analytics_dashboard_export_unknown_resolve_v2"])
	marker := strings.Index(resolve,
		"INSERT INTO platform.analytics_dashboard_export_unknown_resolutions")
	quota := strings.Index(resolve, "UPDATE platform.artifact_quota_counters")
	if marker < 0 || quota < 0 || marker > quota {
		t.Fatal("unknown marker and quota release are not one ordered routine")
	}
	spool := string(fresh["20260802020009_create_analytics_dashboard_export_spool_v1"])
	markerRead := strings.Index(spool,
		"FROM platform.analytics_dashboard_export_unknown_resolutions")
	runningRead := strings.Index(spool, "IF v_job.state='running'")
	if markerRead < 0 || runningRead < 0 || markerRead > runningRead ||
		!strings.Contains(spool[markerRead:runningRead], "RETURN 'inactive'") {
		t.Fatal("committed resolution does not make spool inactive before reaping")
	}
}
