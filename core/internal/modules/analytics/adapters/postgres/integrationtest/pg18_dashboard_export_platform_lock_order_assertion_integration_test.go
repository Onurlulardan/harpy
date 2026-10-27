//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsClaimLoserMadeNoMutation(t *testing.T, ctx context.Context,
	database *sql.DB, beforeRevision int64,
) {
	t.Helper()
	var state, quota, deleteOperation, deleteJob, deleteActor, deleteKey string
	var revision, generation int64
	var platformAttemptMissing bool
	var digest []byte
	err := database.QueryRowContext(ctx, `SELECT state,quota_state,row_revision,
attempt_generation,platform_attempt_id IS NULL,delete_operation_id::text,
delete_job_id::text,delete_actor_id::text,delete_idempotency_key,delete_request_sha256
FROM t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`,
		analyticsLockExport).Scan(&state, &quota, &revision, &generation,
		&platformAttemptMissing, &deleteOperation, &deleteJob, &deleteActor, &deleteKey, &digest)
	if err != nil {
		t.Fatal(err)
	}
	if state != "cancelled" || quota != "released" || revision != beforeRevision+1 ||
		generation != 0 || !platformAttemptMissing || deleteOperation != analyticsLockOperation ||
		deleteJob != analyticsLockDeleteJob || deleteActor != analyticsDashboardActor ||
		deleteKey != "analytics-lock-delete" || !bytes.Equal(digest, bytes.Repeat([]byte{0xd1}, 32)) {
		t.Fatalf("tenant winner-only snapshot differs: %s/%s rev=%d gen=%d op=%s job=%s",
			state, quota, revision, generation, deleteOperation, deleteJob)
	}
	var localAttempts, platformAttempts int
	var jobState, jobError string
	var jobAttempts, used, reserved int64
	err = database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_platform_attempts
 WHERE export_id=$1),(SELECT count(*) FROM platform.job_attempts WHERE job_id=$2),
job.state,job.attempts,job.last_error_code,quota.used_ciphertext_bytes,
quota.reserved_ciphertext_bytes FROM platform.jobs AS job
JOIN platform.artifact_quota_counters AS quota ON quota.tenant_id=job.tenant_id
WHERE job.id=$2`, analyticsLockExport, analyticsLockJob).Scan(&localAttempts,
		&platformAttempts, &jobState, &jobAttempts, &jobError, &used, &reserved)
	if err != nil || localAttempts != 0 || platformAttempts != 0 || jobState != "cancelled" ||
		jobAttempts != 0 || jobError != "EXPORT_CANCELLED" || used != 0 || reserved != 0 {
		t.Fatalf("claim loser mutated evidence: local=%d platform=%d job=%s/%d/%s quota=%d/%d: %v",
			localAttempts, platformAttempts, jobState, jobAttempts, jobError, used, reserved, err)
	}
	var admissionReceipts, runningReceipts, tenantAdmission, tenantRunning int
	var actorAdmission int
	var admissionHeld, runningHeld bool
	err = database.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_capacity_receipts
 WHERE export_id=$1 AND transition_kind='admission_released'),
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_capacity_receipts
 WHERE export_id=$1 AND transition_kind IN ('running_acquired','running_released')),
reservation.admission_held,reservation.running_held,tenant.admission_count,
tenant.running_count,actor.admission_count
FROM t_analytics_a.analytics_dashboard_export_capacity_reservations AS reservation
JOIN t_analytics_a.analytics_dashboard_export_capacity_counters AS tenant
  ON tenant.tenant_id=reservation.tenant_id
JOIN t_analytics_a.analytics_dashboard_export_principal_counters AS actor
  ON actor.actor_id=reservation.actor_id WHERE reservation.export_id=$1`,
		analyticsLockExport).Scan(&admissionReceipts, &runningReceipts, &admissionHeld,
		&runningHeld, &tenantAdmission, &tenantRunning, &actorAdmission)
	if err != nil || admissionReceipts != 1 || runningReceipts != 0 || admissionHeld ||
		runningHeld || tenantAdmission != 0 || tenantRunning != 0 || actorAdmission != 0 {
		t.Fatalf("claim loser mutated capacity: receipts=%d/%d held=%v/%v counters=%d/%d/%d: %v",
			admissionReceipts, runningReceipts, admissionHeld, runningHeld,
			tenantAdmission, tenantRunning, actorAdmission, err)
	}
}
