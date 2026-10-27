//go:build integration

package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertAnalyticsExportUnknownControl(
	t *testing.T, fixture analyticsDashboardExportPG18,
	evidence analyticsDashboardExportUnknownEvidence, expected string,
) {
	t.Helper()
	var state string
	var changedAt time.Time
	err := fixture.runtime.QueryRow(fixture.ctx,
		"SELECT * FROM tenant_a.test_analytics_export_control($1,$2,$3)",
		evidence.exportID, evidence.jobID, time.Now().UTC()).Scan(&state, &changedAt)
	if err != nil || state != expected || changedAt.IsZero() {
		t.Fatalf("unknown control state=%s changed=%v err=%v", state, changedAt, err)
	}
}

func assertAnalyticsExportWriterMoveFence(
	t *testing.T, fixture analyticsDashboardExportPG18,
	evidence analyticsDashboardExportUnknownEvidence,
	cleanupGeneration int64, cleanupOwner uuid.UUID,
) {
	t.Helper()
	if _, err := fixture.database.ExecContext(fixture.ctx, `
UPDATE platform.cutover_groups SET writer='v1' WHERE group_id='analytics';
UPDATE platform.domain_cutovers SET writer='v1' WHERE domain='analytics';`); err != nil {
		t.Fatal(err)
	}
	var leaseUntil time.Time
	var cancelled bool
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_cleanup_checkpoint(
            $1,$2,$3,1,$4,$5,$6,$7,$8,$9)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		cleanupGeneration, cleanupOwner, time.Now().UTC()).Scan(&leaseUntil, &cancelled)
	if analyticsExportSQLState(err) != "42501" {
		t.Fatalf("moved-writer cleanup SQLSTATE=%s err=%v", analyticsExportSQLState(err), err)
	}
	if _, err = fixture.database.ExecContext(fixture.ctx, `
UPDATE platform.cutover_groups SET writer='v2' WHERE group_id='analytics';
UPDATE platform.domain_cutovers SET writer='v2' WHERE domain='analytics';`); err != nil {
		t.Fatal(err)
	}
}

func assertAnalyticsExportResolutionAtomicity(
	t *testing.T, fixture analyticsDashboardExportPG18,
	evidence analyticsDashboardExportUnknownEvidence,
	cleanupGeneration int64, cleanupOwner uuid.UUID, expectedRevision int64,
) {
	t.Helper()
	receiptID := analyticsExportV7(t)
	resolutionDigest := analyticsExportDigest(7)
	resolvedAt := time.Now().UTC()
	transaction, err := fixture.runtime.Begin(fixture.ctx)
	if err != nil {
		t.Fatal(err)
	}
	var resolved bool
	err = transaction.QueryRow(fixture.ctx, `SELECT
        tenant_a.test_analytics_export_unknown_resolve(
            $1,$2,$3,1,$4,$5,$6,$7,$8,$9,'absent',$10,$11)`,
		evidence.exportID, evidence.jobID, evidence.attemptID, evidence.leaseID,
		evidence.effectID, evidence.request, cleanupGeneration, cleanupOwner,
		receiptID, resolutionDigest, resolvedAt).Scan(&resolved)
	if err != nil || !resolved {
		_ = transaction.Rollback(fixture.ctx)
		t.Fatalf("transactional resolve result=%v err=%v", resolved, err)
	}
	if err = transaction.Rollback(fixture.ctx); err != nil {
		t.Fatal(err)
	}
	assertAnalyticsExportResolutionState(
		t, fixture, evidence, 0, 2147483648, expectedRevision, "outcome_unknown",
	)
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT
        tenant_a.test_analytics_export_unknown_resolve(
            $1,$2,$3,1,$4,$5,$6,$7,$8,$9,'absent',$10,$11)`,
		evidence.exportID, evidence.jobID, evidence.attemptID, evidence.leaseID,
		evidence.effectID, evidence.request, cleanupGeneration, cleanupOwner,
		receiptID, resolutionDigest, resolvedAt).Scan(&resolved)
	if err != nil || !resolved {
		t.Fatalf("committed resolve result=%v err=%v", resolved, err)
	}
	assertAnalyticsExportResolutionState(
		t, fixture, evidence, 1, 0, expectedRevision+1, "inactive",
	)
	_, committedRevision := analyticsDashboardExportQuota(t, fixture)
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT
        tenant_a.test_analytics_export_unknown_resolve(
            $1,$2,$3,1,$4,$5,$6,$7,$8,$9,'absent',$10,$11)`,
		evidence.exportID, evidence.jobID, evidence.attemptID, evidence.leaseID,
		evidence.effectID, evidence.request, cleanupGeneration, cleanupOwner,
		receiptID, resolutionDigest, resolvedAt).Scan(&resolved)
	_, replayRevision := analyticsDashboardExportQuota(t, fixture)
	if err != nil || !resolved || replayRevision != committedRevision {
		t.Fatalf("resolve replay result=%v revisions=%d/%d err=%v",
			resolved, committedRevision, replayRevision, err)
	}
	assertAnalyticsExportUnknownControl(t, fixture, evidence, "unknown_resolved")
}

func assertAnalyticsExportResolutionState(
	t *testing.T, fixture analyticsDashboardExportPG18,
	evidence analyticsDashboardExportUnknownEvidence,
	wantMarkers int, wantReserved, wantRevision int64, wantSpool string,
) {
	t.Helper()
	var markers int
	if err := fixture.database.QueryRowContext(fixture.ctx, `SELECT count(*)
        FROM platform.analytics_dashboard_export_unknown_resolutions
        WHERE job_id=$1`, evidence.jobID).Scan(&markers); err != nil {
		t.Fatal(err)
	}
	reserved, revision := analyticsDashboardExportQuota(t, fixture)
	var spool string
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT
        tenant_a.test_analytics_export_spool_state($1,1,$2)`,
		evidence.exportID, evidence.leaseID).Scan(&spool)
	if err != nil || markers != wantMarkers || reserved != wantReserved ||
		revision != wantRevision || spool != wantSpool {
		t.Fatalf("resolution markers=%d reserved=%d revision=%d spool=%s err=%v",
			markers, reserved, revision, spool, err)
	}
}
