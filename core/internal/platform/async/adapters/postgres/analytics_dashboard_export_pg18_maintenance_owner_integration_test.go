//go:build integration

package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestAnalyticsDashboardExportMaintenanceEffectOwnerOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	prepareAnalyticsMaintenancePG18(t, fixture)

	t.Run("active exact and forged effect", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.delete")
		before := analyticsMaintenanceJobSnapshot(t, fixture.ctx, fixture, evidence.jobID)
		var kind string
		err := fixture.runtime.QueryRow(fixture.ctx, `SELECT
tenant_a.test_analytics_export_maintenance_effect_owner($1,$2,$3,$4)`,
			evidence.exportID, evidence.jobID, evidence.effectID, time.Now().UTC()).Scan(&kind)
		if err != nil || kind != evidence.kind {
			t.Fatalf("exact active effect owner kind=%q err=%v", kind, err)
		}
		forged := uuid.NewSHA1(uuid.NameSpaceOID, []byte("forged-maintenance-effect"))
		err = fixture.runtime.QueryRow(fixture.ctx, `SELECT
tenant_a.test_analytics_export_maintenance_effect_owner($1,$2,$3,$4)`,
			evidence.exportID, evidence.jobID, forged, time.Now().UTC()).Scan(&kind)
		after := analyticsMaintenanceJobSnapshot(t, fixture.ctx, fixture, evidence.jobID)
		if analyticsExportSQLState(err) != "P6048" || before != after {
			t.Fatalf("forged effect state=%s changed=%v err=%v",
				analyticsExportSQLState(err), before != after, err)
		}
	})

	t.Run("unknown and succeeded are stale without mutation", func(t *testing.T) {
		unknown := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.delete")
		if _, err := unknown.repository.MarkExternalJobOutcomeUnknown(
			fixture.ctx, unknown.permit,
		); err != nil {
			t.Fatalf("mark direct generic maintenance unknown: %v", err)
		}
		assertMaintenanceOwnerState(t, fixture, unknown, "P6047")
		succeeded := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.expire")
		setAnalyticsMaintenanceTerminal(t, fixture, succeeded, "succeeded")
		assertMaintenanceOwnerState(t, fixture, succeeded, "P6047")
	})

	t.Run("authority race serializes before tenant mutation", func(t *testing.T) {
		evidence := createAnalyticsMaintenanceActive(t, fixture,
			"analytics.dashboard.export.expire")
		runtimeTx, err := fixture.runtime.BeginTx(fixture.ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer runtimeTx.Rollback(fixture.ctx)
		var kind string
		err = runtimeTx.QueryRow(fixture.ctx, `SELECT
tenant_a.test_analytics_export_maintenance_effect_owner($1,$2,$3,$4)`,
			evidence.exportID, evidence.jobID, evidence.effectID, time.Now().UTC()).Scan(&kind)
		if err != nil || kind != evidence.kind {
			t.Fatalf("lock active owner kind=%q err=%v", kind, err)
		}
		moveTx, err := fixture.database.BeginTx(fixture.ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer moveTx.Rollback()
		moveDone := make(chan error, 1)
		go func() {
			_, moveErr := moveTx.ExecContext(fixture.ctx, `UPDATE platform.cutover_groups
SET state='v1',writer='v1',generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='analytics'`, asyncTenantA)
			moveDone <- moveErr
		}()
		select {
		case moveErr := <-moveDone:
			t.Fatalf("authority moved before tenant mutation: %v", moveErr)
		case <-time.After(150 * time.Millisecond):
		}
		if _, err = runtimeTx.Exec(fixture.ctx, `INSERT INTO
tenant_a.test_analytics_export_maintenance_mutations(export_id,kind) VALUES($1,$2)`,
			evidence.exportID, kind); err != nil {
			t.Fatalf("tenant mutation after owner assertion: %v", err)
		}
		if err = runtimeTx.Commit(fixture.ctx); err != nil {
			t.Fatal(err)
		}
		if err = <-moveDone; err != nil {
			t.Fatalf("move authority after tenant commit: %v", err)
		}
		if err = moveTx.Commit(); err != nil {
			t.Fatal(err)
		}
		var mutations int
		if err = fixture.database.QueryRowContext(fixture.ctx, `SELECT count(*) FROM
tenant_a.test_analytics_export_maintenance_mutations WHERE export_id=$1`,
			evidence.exportID).Scan(&mutations); err != nil || mutations != 1 {
			t.Fatalf("serialized tenant mutation count=%d err=%v", mutations, err)
		}
		assertMaintenanceOwnerState(t, fixture, evidence, "42501")
	})
}

func assertMaintenanceOwnerState(
	t *testing.T, fixture analyticsDashboardExportPG18,
	evidence analyticsMaintenanceActive, wantState string,
) {
	t.Helper()
	before := analyticsMaintenanceJobSnapshot(t, fixture.ctx, fixture, evidence.jobID)
	var kind string
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT
tenant_a.test_analytics_export_maintenance_effect_owner($1,$2,$3,$4)`,
		evidence.exportID, evidence.jobID, evidence.effectID, time.Now().UTC()).Scan(&kind)
	after := analyticsMaintenanceJobSnapshot(t, fixture.ctx, fixture, evidence.jobID)
	if analyticsExportSQLState(err) != wantState || before != after {
		t.Fatalf("effect owner state=%s want=%s changed=%v err=%v",
			analyticsExportSQLState(err), wantState, before != after, err)
	}
}
