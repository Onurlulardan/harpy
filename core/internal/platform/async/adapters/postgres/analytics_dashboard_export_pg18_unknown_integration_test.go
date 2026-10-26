//go:build integration

package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAnalyticsDashboardExportUnknownCleanupOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	evidence := createAnalyticsDashboardExportUnknown(t, fixture)
	reserved, revision := analyticsDashboardExportQuota(t, fixture)
	if reserved != 2147483648 {
		t.Fatalf("unknown reserved bytes=%d", reserved)
	}
	unknownSnapshot := analyticsDashboardExportSnapshot(t, fixture, evidence.jobID)
	var replayState string
	var replayAt time.Time
	var replayFailure []byte
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_expired_finalize(
            $1,$2,$3,1,$4,$5,$6,$7)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		evidence.unknownAt).Scan(&replayState, &replayAt, &replayFailure)
	_, replayRevision := analyticsDashboardExportQuota(t, fixture)
	if err != nil || replayState != "outcome_unknown" ||
		!replayAt.Equal(evidence.unknownAt) || len(replayFailure) != 0 ||
		revision != replayRevision ||
		unknownSnapshot != analyticsDashboardExportSnapshot(t, fixture, evidence.jobID) {
		t.Fatalf("unknown replay state=%s revisions=%d/%d err=%v",
			replayState, revision, replayRevision, err)
	}
	if _, err := fixture.database.ExecContext(fixture.ctx,
		"UPDATE platform.tenants SET status='SUSPENDED' WHERE id=$1", asyncTenantA); err != nil {
		t.Fatal(err)
	}
	assertAnalyticsExportUnknownCandidate(t, fixture, evidence)
	assertAnalyticsExportUnknownControl(t, fixture, evidence, "outcome_unknown")
	var storedUnknownAt time.Time
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT
        tenant_a.test_analytics_export_unknown_evidence($1,$2,$3,1,$4,$5,$6)`,
		evidence.exportID, evidence.jobID, evidence.attemptID, evidence.leaseID,
		evidence.effectID, evidence.request).Scan(&storedUnknownAt)
	if err != nil || !storedUnknownAt.Equal(evidence.unknownAt) {
		t.Fatalf("unknown evidence at=%v want=%v err=%v",
			storedUnknownAt, evidence.unknownAt, err)
	}
	if _, err = fixture.database.ExecContext(fixture.ctx, `
UPDATE platform.cutover_groups SET generation=2 WHERE group_id='analytics';
UPDATE platform.domain_cutovers SET generation=2 WHERE domain='analytics';`); err != nil {
		t.Fatal(err)
	}
	cleanupOwner, competingOwner := analyticsExportV7(t), analyticsExportV7(t)
	var claimState string
	var cleanupGeneration int64
	var cleanupUntil time.Time
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_cleanup_claim(
            $1,$2,$3,1,$4,$5,$6,$7,$8)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		cleanupOwner, time.Now().UTC()).Scan(&claimState, &cleanupGeneration, &cleanupUntil)
	if err != nil || claimState != "claimed" || cleanupGeneration != 1 {
		t.Fatalf("cleanup claim state=%s generation=%d err=%v",
			claimState, cleanupGeneration, err)
	}
	var busyGeneration pgtype.Int8
	var busyUntil pgtype.Timestamptz
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_cleanup_claim(
            $1,$2,$3,1,$4,$5,$6,$7,$8)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		competingOwner, time.Now().UTC()).Scan(&claimState, &busyGeneration, &busyUntil)
	if err != nil || claimState != "busy" || busyGeneration.Valid || busyUntil.Valid {
		t.Fatalf("competing cleanup state=%s generation=%+v until=%+v err=%v",
			claimState, busyGeneration, busyUntil, err)
	}
	var cancellationRequested bool
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_cleanup_checkpoint(
            $1,$2,$3,1,$4,$5,$6,$7,$8,$9)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		cleanupGeneration, cleanupOwner, time.Now().UTC()).Scan(
		&cleanupUntil, &cancellationRequested)
	if err != nil || cancellationRequested {
		t.Fatalf("cleanup checkpoint cancelled=%v err=%v", cancellationRequested, err)
	}
	assertAnalyticsExportWriterMoveFence(t, fixture, evidence,
		cleanupGeneration, cleanupOwner)
	assertAnalyticsExportResolutionAtomicity(t, fixture, evidence,
		cleanupGeneration, cleanupOwner, revision)
}

func TestAnalyticsDashboardExportUnknownCleanupDeferOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	evidence := createAnalyticsDashboardExportUnknown(t, fixture)
	owner := analyticsExportV7(t)
	var state string
	var generation int64
	var leaseUntil time.Time
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_cleanup_claim(
            $1,$2,$3,1,$4,$5,$6,$7,$8)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		owner, time.Now().UTC()).Scan(&state, &generation, &leaseUntil)
	if err != nil || state != "claimed" {
		t.Fatalf("defer claim state=%s err=%v", state, err)
	}
	var retryAt time.Time
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT
        tenant_a.test_analytics_export_cleanup_defer(
            $1,$2,$3,1,$4,$5,$6,$7,$8,$9)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		generation, owner, time.Now().UTC()).Scan(&retryAt)
	reserved, _ := analyticsDashboardExportQuota(t, fixture)
	if err != nil || !retryAt.After(time.Now().UTC()) || reserved != 2147483648 {
		t.Fatalf("cleanup defer retry=%v reserved=%d err=%v", retryAt, reserved, err)
	}
	var busyGeneration pgtype.Int8
	var busyUntil pgtype.Timestamptz
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_cleanup_claim(
            $1,$2,$3,1,$4,$5,$6,$7,$8)`, evidence.exportID, evidence.jobID,
		evidence.attemptID, evidence.leaseID, evidence.effectID, evidence.request,
		analyticsExportV7(t), time.Now().UTC()).Scan(&state, &busyGeneration, &busyUntil)
	if err != nil || state != "busy" {
		t.Fatalf("deferred competing claim state=%s err=%v", state, err)
	}
}

func assertAnalyticsExportUnknownCandidate(
	t *testing.T, fixture analyticsDashboardExportPG18,
	evidence analyticsDashboardExportUnknownEvidence,
) {
	t.Helper()
	var tenant, exportID, jobID, attemptID, leaseID, effectID uuid.UUID
	var schema string
	var generation int64
	var request []byte
	var unknownAt time.Time
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        platform.platform_async_analytics_dashboard_export_unknown_candidate_v1()`).Scan(
		&tenant, &schema, &exportID, &jobID, &attemptID, &generation, &leaseID,
		&effectID, &request, &unknownAt)
	if err != nil || tenant != uuid.MustParse(asyncTenantA) || schema != "tenant_a" ||
		exportID != evidence.exportID || jobID != evidence.jobID ||
		attemptID != evidence.attemptID || generation != 1 || leaseID != evidence.leaseID ||
		effectID != evidence.effectID || string(request) != string(evidence.request) ||
		!unknownAt.Equal(evidence.unknownAt) {
		t.Fatalf("unknown candidate job=%s effect=%s at=%v err=%v",
			jobID, effectID, unknownAt, err)
	}
}
