//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAnalyticsDashboardExportLegacyDeleteWorkersAreFencedPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := prepareAnalyticsMaintenanceProtocolPG18(t, ctx)
	connection := analyticsRuntimeConnection(t, ctx, fixture.database)
	baseline := analyticsEffectOwnerTenantSnapshot(t, ctx, fixture.database)
	effectID := fixture.candidate.EffectID
	err := callAnalyticsUnfencedDeleteWorker(ctx, connection, "claim",
		fixture.candidate.ExportID, fixture.candidate.JobID, effectID, fixture.observedAt)
	requireAnalyticsPGCode(t, err, "42501", "legacy unfenced delete claim")
	err = callAnalyticsUnfencedDeleteWorker(ctx, connection, "complete",
		fixture.candidate.ExportID, fixture.candidate.JobID, effectID, fixture.observedAt)
	requireAnalyticsPGCode(t, err, "42501", "legacy unfenced delete complete")
	if after := analyticsEffectOwnerTenantSnapshot(t, ctx, fixture.database); after != baseline {
		t.Fatal("denied legacy delete worker mutated tenant or quota")
	}
}

func TestAnalyticsDashboardExportDeleteWorkersRejectNonOwnerEffectsPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	fixture := prepareAnalyticsMaintenanceProtocolPG18(t, ctx)
	connection := analyticsRuntimeConnection(t, ctx, fixture.database)
	tests := []struct {
		name, exportID, jobID, code string
		effectID                    uuid.UUID
	}{
		{"dead", analyticsMaintenanceDeadExport, analyticsMaintenanceDeadJob,
			"P6047", analyticsMaintenanceEffectID(analyticsMaintenanceDeadJob)},
		{"outcome_unknown", analyticsLockExport, analyticsMaintenanceJob,
			"P6047", fixture.candidate.EffectID},
		{"succeeded", analyticsMaintenanceSuccessExport, analyticsMaintenanceSuccessJob,
			"P6047", analyticsMaintenanceEffectID(analyticsMaintenanceSuccessJob)},
		{"forged", analyticsMaintenanceRunningExport, analyticsMaintenanceRunningJob,
			"P6048", analyticsMaintenanceEffectID("forged-effect-owner")},
	}
	for _, test := range tests {
		for _, worker := range []string{"claim", "complete"} {
			t.Run(test.name+"_"+worker, func(t *testing.T) {
				beforeTenant := analyticsEffectOwnerTenantSnapshot(t, ctx, fixture.database)
				beforePlatform := analyticsEffectOwnerPlatformSnapshot(t, ctx,
					fixture.database, test.jobID)
				err := callAnalyticsFencedDeleteWorker(ctx, connection, worker,
					test.exportID, test.jobID, test.effectID, fixture.observedAt)
				requireAnalyticsPGCode(t, err, test.code, test.name+" "+worker)
				if after := analyticsEffectOwnerTenantSnapshot(t, ctx, fixture.database); after != beforeTenant {
					t.Fatal("rejected effect owner mutated tenant or quota")
				}
				if after := analyticsEffectOwnerPlatformSnapshot(t, ctx,
					fixture.database, test.jobID); after != beforePlatform {
					t.Fatal("rejected effect owner mutated Platform job or attempt")
				}
			})
		}
	}
}

func callAnalyticsUnfencedDeleteWorker(ctx context.Context, connection *sql.Conn,
	worker string, exportID, jobID, effectID uuid.UUID, observedAt time.Time,
) error {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		return err
	}
	if worker == "claim" {
		_, err = transaction.ExecContext(ctx, `SELECT disposition FROM
analytics_dashboard_export_delete_claim_unfenced_v1('t_analytics_a',$1,$2,$3,$4)`,
			exportID, jobID, effectID, observedAt)
		return err
	}
	_, err = transaction.ExecContext(ctx, `SELECT
analytics_dashboard_export_delete_complete_unfenced_v1(
't_analytics_a',$1,$2,$3,1,NULL,NULL,NULL,NULL,$4,$4)`,
		exportID, jobID, effectID, observedAt)
	return err
}

func callAnalyticsFencedDeleteWorker(ctx context.Context, connection *sql.Conn,
	worker, exportID, jobID string, effectID uuid.UUID, observedAt time.Time,
) error {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET LOCAL search_path=t_analytics_a,platform,pg_catalog;
SET LOCAL lock_timeout='10s'; SET LOCAL statement_timeout='20s'`); err != nil {
		return err
	}
	if worker == "claim" {
		_, err = transaction.ExecContext(ctx, `SELECT disposition FROM
analytics_dashboard_export_delete_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
			exportID, jobID, effectID, observedAt)
		return err
	}
	_, err = transaction.ExecContext(ctx, `SELECT analytics_dashboard_export_delete_complete_v1(
't_analytics_a',$1,$2,$3,1,NULL,NULL,NULL,NULL,$4,$4)`,
		exportID, jobID, effectID, observedAt)
	return err
}

func analyticsEffectOwnerTenantSnapshot(t *testing.T, ctx context.Context,
	database *sql.DB,
) string {
	t.Helper()
	var snapshot string
	err := database.QueryRowContext(ctx, `SELECT jsonb_build_object(
'export',(SELECT to_jsonb(item) FROM t_analytics_a.analytics_dashboard_exports item
 WHERE item.export_id=$1),'quota',(SELECT to_jsonb(quota)
 FROM platform.artifact_quota_counters quota WHERE quota.tenant_id=$2))::text`,
		analyticsLockExport, analyticsTenantA).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func analyticsEffectOwnerPlatformSnapshot(t *testing.T, ctx context.Context,
	database *sql.DB, jobID string,
) string {
	t.Helper()
	var snapshot string
	err := database.QueryRowContext(ctx, `SELECT jsonb_build_object(
'job',(SELECT to_jsonb(job) FROM platform.jobs job WHERE job.id=$1),
'attempt',(SELECT to_jsonb(attempt) FROM platform.job_attempts attempt
 WHERE attempt.job_id=$1))::text`, jobID).Scan(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
