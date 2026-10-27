//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsDashboardExportUnknownEvidence(t *testing.T,
	ctx context.Context, database *sql.DB,
	evidence analyticsDashboardExportUnknownEvidence,
) {
	t.Helper()
	var jobExact, platformAttemptExact, exportExact bool
	var tenantAttemptExact, tenantReceiptExact, cardinalityExact bool
	err := database.QueryRowContext(ctx, `SELECT
job.state='dead' AND job.last_error_class IS NULL AND
 job.last_error_code='JOB_OUTCOME_UNKNOWN' AND job.last_error_sha256 IS NULL AND
 job.completed_at=$1 AND job.updated_at=$1 AND job.retain_until=$1+interval '366 days' AND
 job.lease_owner_id IS NULL AND job.lease_heartbeat_at IS NULL AND job.lease_until IS NULL AND
 job.effect_started_at=$2,
platform_attempt.outcome='dead' AND platform_attempt.finished_at=$1 AND
 platform_attempt.error_class IS NULL AND
 platform_attempt.error_code='JOB_OUTCOME_UNKNOWN' AND
 platform_attempt.redacted_error_sha256 IS NULL AND
 platform_attempt.effect_receipt_id IS NULL AND
 platform_attempt.effect_started_at=$2 AND platform_attempt.effect_id=$4 AND
 platform_attempt.effect_request_sha256=encode($5::bytea,'hex'),
export.state='running' AND export.quota_state='held' AND
 export.failure_code='ARTIFACT_OUTCOME_UNKNOWN' AND export.updated_at=$1 AND
 export.lease_owner IS NULL AND export.lease_until IS NULL AND
 export.platform_attempt_id=$3 AND export.platform_effect_id=$4 AND
 export.platform_effect_request_sha256=$5 AND export.attempt_generation=1,
tenant_attempt.outcome='unknown' AND tenant_attempt.failure_class IS NULL AND
 tenant_attempt.failure_code='JOB_OUTCOME_UNKNOWN' AND
 tenant_attempt.failure_sha256 IS NULL AND tenant_attempt.finished_at=$1 AND
 tenant_attempt.platform_attempt_id=$3 AND tenant_attempt.lease_owner=$6 AND
 tenant_attempt.platform_effect_id=$4 AND
 tenant_attempt.platform_effect_request_sha256=$5 AND
 tenant_attempt.resolution_receipt_id IS NULL,
tenant_receipt.outcome='unknown' AND tenant_receipt.completed_at=$1 AND
 tenant_receipt.platform_attempt_id=$3 AND tenant_receipt.platform_effect_id=$4 AND
 tenant_receipt.platform_effect_request_sha256=$5,
(SELECT count(*)=1 FROM platform.job_attempts WHERE job_id=$7) AND
 (SELECT count(*)=1 FROM t_analytics_a.analytics_dashboard_export_platform_attempts
  WHERE export_id=$8) AND
 (SELECT count(*)=1 FROM t_analytics_a.analytics_dashboard_export_attempt_receipts
  WHERE export_id=$8)
FROM platform.jobs job
JOIN platform.job_attempts platform_attempt ON platform_attempt.job_id=job.id
JOIN t_analytics_a.analytics_dashboard_exports export ON export.job_id=job.id
JOIN t_analytics_a.analytics_dashboard_export_platform_attempts tenant_attempt
 ON tenant_attempt.export_id=export.export_id
JOIN t_analytics_a.analytics_dashboard_export_attempt_receipts tenant_receipt
 ON tenant_receipt.export_id=export.export_id
 AND tenant_receipt.attempt_generation=tenant_attempt.attempt_generation
WHERE job.id=$7 AND export.export_id=$8 AND platform_attempt.attempt_id=$3`,
		evidence.unknownAt, evidence.effect.startedAt, evidence.claim.attemptID,
		evidence.effect.effectID, evidence.effect.request, analyticsUnknownLeaseV2,
		analyticsUnknownJobV2, analyticsUnknownExportV2).Scan(&jobExact,
		&platformAttemptExact, &exportExact, &tenantAttemptExact,
		&tenantReceiptExact, &cardinalityExact)
	if err != nil || !jobExact || !platformAttemptExact || !exportExact ||
		!tenantAttemptExact || !tenantReceiptExact || !cardinalityExact {
		t.Fatalf("render outcome_unknown exact evidence = %v/%v/%v/%v/%v/%v: %v",
			jobExact, platformAttemptExact, exportExact, tenantAttemptExact,
			tenantReceiptExact, cardinalityExact, err)
	}
}

func analyticsDashboardExportUnknownSnapshot(t *testing.T, ctx context.Context,
	database *sql.DB,
) string {
	t.Helper()
	lifecycle := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsUnknownExportV2, analyticsUnknownJobV2)
	var recovery string
	err := database.QueryRowContext(ctx, `SELECT jsonb_build_object(
'cleanup_claim',(SELECT to_jsonb(item) FROM
 platform.analytics_dashboard_export_unknown_cleanup_claims item WHERE item.job_id=$1),
'platform_resolution',(SELECT to_jsonb(item) FROM
 platform.analytics_dashboard_export_unknown_resolutions item WHERE item.job_id=$1),
'tenant_resolution',(SELECT to_jsonb(item) FROM
 t_analytics_a.analytics_dashboard_export_unknown_resolutions item
 WHERE item.export_id=$2))::text`, analyticsUnknownJobV2,
		analyticsUnknownExportV2).Scan(&recovery)
	if err != nil {
		t.Fatal(err)
	}
	return lifecycle + recovery
}
