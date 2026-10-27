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

func TestAnalyticsDashboardExportCleanupLeaseContract(t *testing.T) {
	claim := analyticsExportFresh(t,
		"20260802020015_create_analytics_dashboard_export_unknown_cleanup_claim_v1")
	for _, token := range []string{"claim_state text", "cleanup_generation bigint",
		"lease_owner_id=p_cleanup_lease_owner_id", "claim_state:='busy'",
		"lease_until=v_claimed_at+interval '60 seconds'"} {
		if !strings.Contains(claim, token) {
			t.Fatalf("cleanup claim misses %q", token)
		}
	}
	checkpoint := analyticsExportFresh(t,
		"20260802020016_create_analytics_dashboard_export_unknown_cleanup_checkpoint_v1")
	deferRoutine := analyticsExportFresh(t,
		"20260802020017_create_analytics_dashboard_export_unknown_cleanup_defer_v1")
	for name, document := range map[string]string{
		"checkpoint": checkpoint, "defer": deferRoutine,
	} {
		for _, token := range []string{"p_cleanup_generation", "p_cleanup_lease_owner_id",
			"lease_until<=pg_catalog.clock_timestamp()", "P6047"} {
			if !strings.Contains(document, token) {
				t.Fatalf("cleanup %s misses %q", name, token)
			}
		}
	}
	retire := analyticsExportFresh(t,
		"20260802020023_fence_analytics_dashboard_export_unknown_resolve_v1")
	if !strings.Contains(retire, "cleanup lease is required") {
		t.Fatal("legacy unknown resolution ABI remains writable")
	}
}

func TestAnalyticsDashboardExportDedicatedExpiryAndControlContracts(t *testing.T) {
	candidate := analyticsExportFresh(t,
		"20260802020019_create_analytics_dashboard_export_expired_candidate_v1")
	finalize := analyticsExportFresh(t,
		"20260802020020_create_analytics_dashboard_export_expired_finalize_v1")
	control := analyticsExportFresh(t,
		"20260802020021_create_analytics_dashboard_export_control_v1")
	for _, token := range []string{"effect_started boolean", "tenant.deleted_at IS NULL",
		"job.lease_until<=pg_catalog.clock_timestamp()", "LIMIT 1"} {
		if !strings.Contains(candidate, token) {
			t.Fatalf("expired candidate misses %q", token)
		}
	}
	for _, token := range []string{"outcome_unknown", "JOB_LEASE_EXPIRED",
		"reserved_ciphertext_bytes-2147483648", "v_attempt.effect_started_at IS NOT NULL"} {
		if !strings.Contains(finalize, token) {
			t.Fatalf("expired finalizer misses %q", token)
		}
	}
	for _, token := range []string{"cancellation_requested", "outcome_unknown",
		"unknown_resolved", "EXPORT_CANCELLED", "FOR UPDATE"} {
		if !strings.Contains(control, token) {
			t.Fatalf("control routine misses %q", token)
		}
	}
	guard := analyticsExportFresh(t,
		"20260802020025_allow_analytics_dashboard_export_current_unknown_v1")
	for _, token := range []string{"analytics.dashboard.export.render",
		"platform_async_guard_job_attempt_v3", "assert_cleanup_owner_v1",
		"platform_async_validate_stale_external_attempt_unknown_v1"} {
		if !strings.Contains(guard, token) {
			t.Fatalf("Analytics-specific unknown guard misses %q", token)
		}
	}
	jobGuard := analyticsExportFresh(t,
		"20260802020026_allow_analytics_dashboard_export_current_unknown_job_v1")
	for _, token := range []string{"analytics.dashboard.export.render",
		"platform_async_guard_job_transition_v3", "assert_cleanup_owner_v1",
		"platform_async_validate_stale_external_job_unknown_v1"} {
		if !strings.Contains(jobGuard, token) {
			t.Fatalf("Analytics-specific unknown job guard misses %q", token)
		}
	}
}

func TestAnalyticsDashboardExportRuntimeGrantsStayReadOnly(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	var grants strings.Builder
	for _, pair := range pairs {
		grants.Write(pair["fresh"])
	}
	granted := grants.String()
	for _, forbidden := range []string{"export_checkpoint_v1(\n", "export_effect_start_v1(\n",
		"expired_finalize_v1(\n", "unknown_cleanup_claim_v1(\n", "export_control_v1(\n"} {
		if strings.Contains(granted, "GRANT EXECUTE ON FUNCTION\n    platform.platform_async_analytics_dashboard_"+forbidden) {
			t.Fatalf("owner-only routine received direct runtime grant: %s", forbidden)
		}
	}
	for _, required := range []string{"export_candidate_v1()", "export_unknown_candidate_v1()",
		"export_expired_candidate_v1()", "export_spool_tenant_v1(uuid)"} {
		if !strings.Contains(granted, required) {
			t.Fatalf("read-only worker routine is not granted: %s", required)
		}
	}
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
