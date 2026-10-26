//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type analyticsMaintenanceRaceResult struct {
	state string
	err   error
}

func TestAnalyticsDashboardExportMaintenanceWinsDeleteWorkerRacePG18(t *testing.T) {
	for _, worker := range []string{"claim", "complete"} {
		t.Run(worker, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := prepareAnalyticsMaintenanceProtocolPG18(t, ctx)
			claim := claimAnalyticsMaintenanceUnknown(t, ctx, fixture)
			var beforeRevision int64
			if err := fixture.database.QueryRowContext(ctx, `SELECT row_revision FROM
t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`,
				claim.candidate.ExportID).Scan(&beforeRevision); err != nil {
				t.Fatal(err)
			}
			blocker, err := fixture.database.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer blocker.Close()
			blockingTx, err := blocker.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = blockingTx.Rollback() }()
			if _, err = blockingTx.ExecContext(ctx,
				`SELECT 1 FROM platform.jobs WHERE id=$1 FOR UPDATE`,
				claim.candidate.JobID); err != nil {
				t.Fatal(err)
			}
			maintenanceConnection := analyticsRuntimeConnection(t, ctx, fixture.database)
			loserConnection := analyticsRuntimeConnection(t, ctx, fixture.database)
			setAnalyticsRaceApplicationName(t, ctx, maintenanceConnection,
				"analytics-maintenance-winner-"+worker)
			setAnalyticsRaceApplicationName(t, ctx, loserConnection,
				"analytics-delete-loser-"+worker)
			winnerResult := make(chan analyticsMaintenanceRaceResult, 1)
			loserResult := make(chan error, 1)
			go func() {
				state, resolveErr := runAnalyticsMaintenanceResolve(ctx,
					maintenanceConnection, claim, claim.effectReceiptID,
					claim.effectReceiptSHA, claim.resolutionSHA256)
				winnerResult <- analyticsMaintenanceRaceResult{state: state, err: resolveErr}
			}()
			waitAnalyticsRaceLock(t, ctx, fixture.database,
				"analytics-maintenance-winner-"+worker)
			go func() {
				loserResult <- callAnalyticsFencedDeleteWorker(ctx, loserConnection, worker,
					claim.candidate.ExportID.String(), claim.candidate.JobID.String(),
					claim.candidate.EffectID, claim.resolvedAt)
			}()
			waitAnalyticsRaceLock(t, ctx, fixture.database,
				"analytics-delete-loser-"+worker)
			if err = blockingTx.Commit(); err != nil {
				t.Fatal(err)
			}
			winner := <-winnerResult
			if winner.err != nil || winner.state != "deleted" {
				t.Fatalf("maintenance winner = %s: %v", winner.state, winner.err)
			}
			loserErr := <-loserResult
			requireAnalyticsPGCode(t, loserErr, "P6047",
				"maintenance versus delete "+worker+" loser")
			assertAnalyticsMaintenanceWinnerOnly(t, ctx, fixture.database,
				claim, beforeRevision)
		})
	}
}

func assertAnalyticsMaintenanceWinnerOnly(t *testing.T, ctx context.Context,
	database *sql.DB, claim analyticsMaintenanceClaimFixture, beforeRevision int64,
) {
	t.Helper()
	var tenantState, quotaState, jobState, attemptOutcome string
	var revision, used, reserved int64
	var maintenanceCleared bool
	var receiptID string
	var platformReceipts, platformResolutions, tenantResolutions int
	err := database.QueryRowContext(ctx, `SELECT item.state,item.quota_state,item.row_revision,
item.maintenance_unknown_job_id IS NULL AND item.maintenance_cleanup_generation IS NULL
AND item.maintenance_cleanup_lease_owner IS NULL,
quota.used_ciphertext_bytes,quota.reserved_ciphertext_bytes,job.state,attempt.outcome,
attempt.effect_receipt_id::text,
(SELECT count(*) FROM platform.idempotency_receipts WHERE causation_id=job.id),
(SELECT count(*) FROM platform.analytics_dashboard_export_maintenance_unknown_resolutions
 WHERE job_id=job.id),(SELECT count(*) FROM
 t_analytics_a.analytics_dashboard_export_maintenance_unknown_resolutions
 WHERE job_id=job.id)
FROM t_analytics_a.analytics_dashboard_exports item
JOIN platform.artifact_quota_counters quota ON quota.tenant_id=item.tenant_id
JOIN platform.jobs job ON job.id=item.delete_job_id
JOIN platform.job_attempts attempt ON attempt.job_id=job.id
WHERE item.export_id=$1`, claim.candidate.ExportID).Scan(&tenantState, &quotaState,
		&revision, &maintenanceCleared, &used, &reserved, &jobState, &attemptOutcome,
		&receiptID, &platformReceipts, &platformResolutions, &tenantResolutions)
	if err != nil || tenantState != "deleted" || quotaState != "released" ||
		revision != beforeRevision+1 || !maintenanceCleared || used != 0 || reserved != 0 ||
		jobState != "succeeded" || attemptOutcome != "succeeded" ||
		receiptID != claim.effectReceiptID.String() || platformReceipts != 1 ||
		platformResolutions != 1 || tenantResolutions != 1 {
		t.Fatalf("maintenance winner-only evidence differs: tenant=%s/%s rev=%d "+
			"job=%s/%s receipt=%s counts=%d/%d/%d quota=%d/%d: %v",
			tenantState, quotaState, revision, jobState, attemptOutcome, receiptID,
			platformReceipts, platformResolutions, tenantResolutions, used, reserved, err)
	}
}
