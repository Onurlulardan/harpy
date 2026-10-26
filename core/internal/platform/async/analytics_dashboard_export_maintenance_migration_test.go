package async

import (
	"strings"
	"testing"
)

func TestAnalyticsDashboardExportMaintenanceUnknownContract(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	candidate := string(pairs["20260802020030_create_analytics_dashboard_export_maintenance_unknown_candidate_v1"]["fresh"])
	for _, token := range []string{
		"analytics.dashboard.export.delete", "analytics.dashboard.export.expire",
		"attempt.lease_generation", "attempt.effect_id)=5", "tenant.schema_name",
		"authority_group_generation", "cutover_generation", "resolution.job_id=job.id",
	} {
		if !strings.Contains(candidate, token) {
			t.Fatalf("maintenance candidate misses %q", token)
		}
	}
	claim := string(pairs["20260802020031_create_analytics_dashboard_export_maintenance_unknown_claim_v1"]["fresh"])
	for _, token := range []string{
		"claim_state:='resolved'", "v_resolution.original_unknown_at<>p_unknown_at",
		"v_job.state<>'succeeded'", "v_attempt.effect_receipt_id<>",
		"claim_state:='busy'", "P6046", "P6048",
	} {
		if !strings.Contains(claim, token) {
			t.Fatalf("maintenance claim misses %q", token)
		}
	}
}

func TestAnalyticsDashboardExportMaintenanceResolutionContract(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	resolve := string(pairs["20260802020039_create_analytics_dashboard_export_maintenance_unknown_resolve_v1"]["fresh"])
	for _, token := range []string{
		"maintenance_assert_replay_v1", "maintenance_insert_receipt_v1",
		"original_attempt_outcome", "effect_receipt_id=p_effect_receipt_id",
		"state='succeeded'", "retain_until=p_resolved_at+interval '90 days'",
		"claim.authority_group_generation=p_authority_group_generation",
		"claim.unknown_at=p_unknown_at", "P6046", "P6048",
	} {
		if !strings.Contains(resolve, token) {
			t.Fatalf("maintenance resolve misses %q", token)
		}
	}
	if strings.Contains(resolve, "artifact_quota_counters") {
		t.Fatal("Platform maintenance resolution mutates tenant-owned quota")
	}
	helpers := string(pairs["20260802020038_create_analytics_dashboard_export_maintenance_resolution_helpers_v1"]["fresh"])
	if strings.Index(helpers, "INSERT INTO platform.idempotency_receipts") < 0 ||
		!strings.Contains(helpers, "maintenance_receipt_sha256_v1") {
		t.Fatal("maintenance effect receipt is not inserted with canonical evidence")
	}
}

func TestAnalyticsDashboardExportMaintenanceEffectOwnerContract(t *testing.T) {
	owner := string(analyticsDashboardExportMigrationPairs(t)["20260802020041_create_analytics_dashboard_export_maintenance_effect_owner_v1"]["fresh"])
	jobLock := strings.Index(owner, "WHERE job.id=p_job_id FOR UPDATE")
	attemptLock := strings.Index(owner, "FROM platform.job_attempts attempt")
	if jobLock < 0 || attemptLock < 0 || jobLock > attemptLock {
		t.Fatal("maintenance effect owner does not lock Platform job first")
	}
	for _, token := range []string{
		"uuid_extract_version(p_effect_id) IS DISTINCT FROM 5",
		"v_job.state IS DISTINCT FROM 'running'", "v_attempt.effect_id IS DISTINCT FROM",
		"group_row.generation=v_job.authority_group_generation",
		"domain_row.generation=v_job.cutover_generation", "P6047", "P6048", "42501",
	} {
		if !strings.Contains(owner, token) {
			t.Fatalf("maintenance effect owner misses %q", token)
		}
	}
}

func TestAnalyticsDashboardExportMaintenanceRuntimeGrantIsCandidateOnly(t *testing.T) {
	grant := string(analyticsDashboardExportMigrationPairs(t)["20260802020040_grant_analytics_dashboard_export_maintenance_unknown_candidate_v1"]["fresh"])
	if !strings.Contains(grant, "maintenance_unknown_candidate_v1()") ||
		strings.Contains(grant, "maintenance_unknown_claim_v1") ||
		strings.Contains(grant, "maintenance_unknown_resolve_v1") {
		t.Fatal("maintenance runtime grant is not read-only candidate access")
	}
}

func TestAnalyticsDashboardExportMaintenanceGenericUnknownGuardsAreCallable(t *testing.T) {
	grant := string(analyticsDashboardExportMigrationPairs(t)["20260802020042_grant_analytics_dashboard_export_maintenance_unknown_guards_v1"]["fresh"])
	for _, token := range []string{
		"maintenance_unknown_attempt_v1", "maintenance_unknown_job_v1",
		"TO harpy_runtime",
	} {
		if !strings.Contains(grant, token) {
			t.Fatalf("generic maintenance unknown guard grant misses %q", token)
		}
	}
	if strings.Contains(grant, "maintenance_unknown_resolve_v1") ||
		strings.Contains(grant, "maintenance_effect_owner_v1") {
		t.Fatal("owner-only maintenance mutation API received a direct runtime grant")
	}
}

func TestAnalyticsDashboardExportMaintenanceReceiptIdentityIsDerivedInSQL(t *testing.T) {
	pairs := analyticsDashboardExportMigrationPairs(t)
	identity := string(pairs["20260802020043_create_analytics_dashboard_export_maintenance_receipt_identity_v1"]["fresh"])
	binding := string(pairs["20260802020044_bind_analytics_dashboard_export_maintenance_receipt_identity_v1"]["fresh"])
	replayBinding := string(pairs["20260802020045_bind_analytics_dashboard_export_maintenance_replay_identity_v1"]["fresh"])
	for _, token := range []string{
		"8e7a6d02-df6d-4327-b179-160296b50b8e", "decode('00','hex')",
		"get_byte(v_uuid,6)&15)|80", "get_byte(v_uuid,8)&63)|128",
		"platform_async_analytics_export_sha1_v1",
	} {
		if !strings.Contains(identity, token) {
			t.Fatalf("SQL receipt identity derivation misses %q", token)
		}
	}
	for _, logical := range []string{
		"20260802020034_create_analytics_dashboard_export_maintenance_receipt_evidence_v1",
		"20260802020038_create_analytics_dashboard_export_maintenance_resolution_helpers_v1",
	} {
		if strings.Contains(string(pairs[logical]["fresh"]),
			"platform_async_analytics_export_maint_receipt_id_v1") {
			t.Fatalf("%s depends on a later migration", logical)
		}
	}
	for _, token := range []string{
		"CREATE OR REPLACE FUNCTION platform.platform_async_assert_analytics_dashboard_export_maintenance_receipt_v1",
		"CREATE OR REPLACE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_insert_receipt_v1",
		"platform_async_analytics_export_maint_receipt_id_v1",
	} {
		if !strings.Contains(binding, token) {
			t.Fatalf("receipt identity binding misses %q", token)
		}
	}
	for _, token := range []string{
		"CREATE OR REPLACE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_assert_replay_v1",
		"p_effect_receipt_id IS DISTINCT FROM v_expected_receipt_id",
		"p_resolution.effect_receipt_id IS DISTINCT FROM v_expected_receipt_id",
		"v_receipt.receipt_id IS DISTINCT FROM v_expected_receipt_id",
	} {
		if !strings.Contains(replayBinding, token) {
			t.Fatalf("receipt replay identity binding misses %q", token)
		}
	}
}
