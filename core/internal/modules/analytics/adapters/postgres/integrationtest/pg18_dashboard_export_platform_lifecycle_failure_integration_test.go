//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsDashboardExportFailureLifecycle(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	claim := claimAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsFailureExport, analyticsFailureJob, analyticsFailureLeaseA, 1)
	assertAnalyticsExportQuota(t, ctx, database, 1024, 2147483648)
	finishedAt := claim.claimedAt.Add(time.Millisecond)
	state, receipt := finishAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsFailureExport, analyticsFailureJob, analyticsFailureLeaseA,
		claim, "", nil, 1, "transient", "UPLOAD_RETRY",
		true, "provider-retry-1", finishedAt)
	if state != "queued" || receipt.Valid {
		t.Fatalf("retryable Platform finish = %s/%v", state, receipt)
	}
	after := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsFailureExport, analyticsFailureJob)
	replayState, replayReceipt := finishAnalyticsDashboardExportPlatform(t, ctx,
		connection, analyticsFailureExport, analyticsFailureJob,
		analyticsFailureLeaseA, claim, "", nil, 1,
		"transient", "UPLOAD_RETRY", true, "provider-retry-1", finishedAt)
	if replayState != state || replayReceipt != receipt || after !=
		analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
			analyticsFailureExport, analyticsFailureJob) {
		t.Fatal("retryable Platform finish replay changed exact evidence")
	}
	assertAnalyticsDashboardExportFailurePersisted(t, ctx, database, 1,
		"queued", "none", 4, "queued", "requeued", "transient",
		"UPLOAD_RETRY", false, true, 2, 3)
	assertAnalyticsExportQuota(t, ctx, database, 1024, 0)

	var retryAt time.Time
	if err := database.QueryRowContext(ctx,
		`SELECT run_at FROM platform.jobs WHERE id=$1`, analyticsFailureJob,
	).Scan(&retryAt); err != nil {
		t.Fatal(err)
	}
	waitUntilAnalyticsDashboardExportExpiry(t, ctx, retryAt)
	claim = claimAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsFailureExport, analyticsFailureJob, analyticsFailureLeaseB, 2)
	assertAnalyticsExportQuota(t, ctx, database, 1024, 2147483648)
	effect := startAnalyticsDashboardExportPlatformEffect(t, ctx, connection,
		analyticsFailureExport, analyticsFailureJob, analyticsFailureLeaseB,
		analyticsFailureEffect, claim, 2, 0xc1, claim.claimedAt.Add(time.Millisecond))
	state, receipt = finishAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsFailureExport, analyticsFailureJob, analyticsFailureLeaseB,
		claim, effect.effectID, effect.request, 2, "permanent", "UPLOAD_REJECTED",
		false, "provider-rejected-2", effect.startedAt.Add(time.Millisecond))
	if state != "failed" || !receipt.Valid {
		t.Fatalf("terminal Platform finish = %s/%v", state, receipt)
	}
	assertAnalyticsDashboardExportFailurePersisted(t, ctx, database, 2,
		"failed", "released", 7, "dead", "dead", "permanent",
		"UPLOAD_REJECTED", true, false, 1, 6)
	assertAnalyticsExportQuota(t, ctx, database, 1024, 0)
}

func assertAnalyticsDashboardExportFailurePersisted(t *testing.T,
	ctx context.Context, database *sql.DB, generation int64, tenantState,
	quotaState string, revision int64, jobState, outcome, errorClass, errorCode string,
	receiptPresent, admissionHeld bool, admissionCount, receiptCount int64,
) {
	t.Helper()
	var gotTenantState, gotQuota, gotJobState, gotOutcome, gotClass, gotCode string
	var gotRevision, gotGeneration, jobAttempts int64
	var gotReceipt, gotAdmission, gotRunning bool
	var tenantAdmission, tenantRunning, principalAdmission, receipts int64
	var admissionAcquired, admissionReleased, runningAcquired, runningReleased int64
	err := database.QueryRowContext(ctx, `SELECT item.state,item.quota_state,
item.row_revision,item.attempt_generation,job.state,job.attempts,attempt.outcome,
attempt.error_class,attempt.error_code,attempt.effect_receipt_id IS NOT NULL,
reservation.admission_held,reservation.running_held,tenant.admission_count,
tenant.running_count,principal.admission_count,
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_capacity_receipts
 WHERE export_id=item.export_id),
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_capacity_receipts
 WHERE export_id=item.export_id AND transition_kind='admission_acquired'),
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_capacity_receipts
 WHERE export_id=item.export_id AND transition_kind='admission_released'),
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_capacity_receipts
 WHERE export_id=item.export_id AND transition_kind='running_acquired'),
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_export_capacity_receipts
 WHERE export_id=item.export_id AND transition_kind='running_released')
FROM t_analytics_a.analytics_dashboard_exports item
JOIN platform.jobs job ON job.id=item.job_id
JOIN platform.job_attempts attempt ON attempt.job_id=job.id AND attempt.attempt=$2
JOIN t_analytics_a.analytics_dashboard_export_capacity_reservations reservation
 ON reservation.export_id=item.export_id
JOIN t_analytics_a.analytics_dashboard_export_capacity_counters tenant
 ON tenant.tenant_id=item.tenant_id
JOIN t_analytics_a.analytics_dashboard_export_principal_counters principal
 ON principal.actor_id=item.actor_id
WHERE item.export_id=$1`, analyticsFailureExport, generation).Scan(&gotTenantState,
		&gotQuota, &gotRevision, &gotGeneration, &gotJobState, &jobAttempts,
		&gotOutcome, &gotClass, &gotCode, &gotReceipt, &gotAdmission, &gotRunning,
		&tenantAdmission, &tenantRunning, &principalAdmission, &receipts,
		&admissionAcquired, &admissionReleased, &runningAcquired, &runningReleased)
	if err != nil || gotTenantState != tenantState || gotQuota != quotaState ||
		gotRevision != revision || gotGeneration != generation ||
		gotJobState != jobState || jobAttempts != generation || gotOutcome != outcome ||
		gotClass != errorClass || gotCode != errorCode || gotReceipt != receiptPresent ||
		gotAdmission != admissionHeld || gotRunning || tenantAdmission != admissionCount ||
		tenantRunning != 0 || principalAdmission != admissionCount || receipts != receiptCount ||
		admissionAcquired != 1 || admissionReleased != generation-1 ||
		runningAcquired != generation || runningReleased != generation {
		t.Fatalf("failure persistence gen=%d tenant=%s/%s/rev%d job=%s/%d "+
			"attempt=%s/%s/%s/receipt=%v capacity=%v/%v/%d/%d/%d receipts=%d: %v",
			gotGeneration, gotTenantState, gotQuota, gotRevision, gotJobState,
			jobAttempts, gotOutcome, gotClass, gotCode, gotReceipt, gotAdmission,
			gotRunning, tenantAdmission, tenantRunning, principalAdmission, receipts, err)
	}
}

func finishAnalyticsDashboardExportPlatform(t *testing.T, ctx context.Context,
	connection *sql.Conn, exportID, jobID, leaseOwner string,
	claim analyticsPlatformExportClaim, effectID string, request []byte,
	generation int64, failureClass, failureCode string, retryable bool,
	providerReference string, finishedAt time.Time,
) (string, sql.NullString) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var effectArgument, requestArgument, providerArgument any
	if effectID != "" {
		effectArgument, requestArgument = effectID, request
	}
	if providerReference != "" {
		providerArgument = providerReference
	}
	var state string
	var receipt sql.NullString
	err := transaction.QueryRowContext(ctx, `SELECT state,effect_receipt_id FROM
analytics_dashboard_export_platform_finish_v1(
't_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		analyticsTenantA, exportID, jobID, claim.attemptID, generation, leaseOwner,
		effectArgument, requestArgument, failureClass, failureCode,
		bytes.Repeat([]byte{0xd1}, 32), retryable, providerArgument, finishedAt,
	).Scan(&state, &receipt)
	if err != nil {
		t.Fatalf("Platform-first finish %s generation %d: %v", exportID, generation, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return state, receipt
}
