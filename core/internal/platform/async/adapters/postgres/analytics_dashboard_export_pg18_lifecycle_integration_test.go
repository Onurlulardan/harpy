//go:build integration

package postgres

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAnalyticsDashboardExportACLAndPreEffectRecoveryOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	if uuid.MustParse(asyncTenantA).Version() != uuid.Version(4) {
		t.Fatal("fixture tenant must prove UUIDv4 tenant support")
	}
	exportID, jobID := analyticsExportV7(t), analyticsExportV7(t)
	leaseID := analyticsExportV7(t)
	seedAnalyticsDashboardExportJob(
		t, fixture.ctx, fixture.runtime, exportID, jobID, time.Now().UTC().Add(-time.Second),
	)
	claim := claimAnalyticsDashboardExport(
		t, fixture.ctx, fixture.runtime, exportID, jobID, leaseID, 1,
	)
	if claim.state != "claimed" {
		t.Fatalf("claim state=%s", claim.state)
	}
	attemptID := analyticsExportAttemptID(t, claim)
	assertAnalyticsExportDirectMutationDenied(
		t, fixture, exportID, jobID, attemptID, leaseID,
	)
	setAnalyticsDashboardExportLease(
		t, fixture, jobID, time.Now().UTC().Add(-time.Second),
	)
	assertAnalyticsExportExpiredCandidate(
		t, fixture, exportID, jobID, attemptID, leaseID, uuid.Nil, nil,
	)
	assertGenericRecoverySkipsAnalyticsExport(t, fixture, jobID)
	observedAt := time.Now().UTC()
	var state string
	var finalized time.Time
	var failure []byte
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_expired_finalize(
            $1,$2,$3,1,$4,NULL,NULL,$5)`,
		exportID, jobID, attemptID, leaseID, observedAt,
	).Scan(&state, &finalized, &failure)
	if err != nil || state != "queued" || len(failure) != 32 {
		t.Fatalf("pre-effect expiry state=%s digest=%d err=%v", state, len(failure), err)
	}
	reserved, revision := analyticsDashboardExportQuota(t, fixture)
	firstState, firstFinalized := state, finalized
	firstFailure := append([]byte(nil), failure...)
	finalizedSnapshot := analyticsDashboardExportSnapshot(t, fixture, jobID)
	if reserved != 0 {
		t.Fatalf("pre-effect expiry reserved bytes=%d", reserved)
	}
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_expired_finalize(
            $1,$2,$3,1,$4,NULL,NULL,$5)`,
		exportID, jobID, attemptID, leaseID, observedAt,
	).Scan(&state, &finalized, &failure)
	_, replayRevision := analyticsDashboardExportQuota(t, fixture)
	if err != nil || state != "queued" || replayRevision != revision {
		t.Fatalf("expiry replay state=%s revisions=%d/%d err=%v",
			state, revision, replayRevision, err)
	}
	if state != firstState || !finalized.Equal(firstFinalized) ||
		!bytes.Equal(failure, firstFailure) ||
		analyticsDashboardExportSnapshot(t, fixture, jobID) != finalizedSnapshot {
		t.Fatal("expiry replay changed exact job, attempt, failure, or finalization evidence")
	}
}

func TestAnalyticsDashboardExportCancelControlOnPG184(t *testing.T) {
	fixture := startAnalyticsDashboardExportPG18(t)
	queuedExport, queuedJob := analyticsExportV7(t), analyticsExportV7(t)
	seedAnalyticsDashboardExportJob(t, fixture.ctx, fixture.runtime,
		queuedExport, queuedJob, time.Now().UTC().Add(-time.Second))
	requestedAt := time.Now().UTC()
	var state string
	var changedAt time.Time
	err := fixture.runtime.QueryRow(fixture.ctx,
		"SELECT * FROM tenant_a.test_analytics_export_control($1,$2,$3)",
		queuedExport, queuedJob, requestedAt).Scan(&state, &changedAt)
	if err != nil || state != "cancelled" {
		t.Fatalf("queued cancel state=%s err=%v", state, err)
	}
	var replayAt time.Time
	err = fixture.runtime.QueryRow(fixture.ctx,
		"SELECT * FROM tenant_a.test_analytics_export_control($1,$2,$3)",
		queuedExport, queuedJob, requestedAt).Scan(&state, &replayAt)
	if err != nil || state != "cancelled" || !replayAt.Equal(changedAt) {
		t.Fatalf("queued cancel replay state=%s changed=%v/%v err=%v",
			state, changedAt, replayAt, err)
	}
	runningExport, runningJob := analyticsExportV7(t), analyticsExportV7(t)
	runningLease := analyticsExportV7(t)
	seedAnalyticsDashboardExportJob(t, fixture.ctx, fixture.runtime,
		runningExport, runningJob, time.Now().UTC().Add(-time.Second))
	claim := claimAnalyticsDashboardExport(t, fixture.ctx, fixture.runtime,
		runningExport, runningJob, runningLease, 1)
	attemptID := analyticsExportAttemptID(t, claim)
	err = fixture.runtime.QueryRow(fixture.ctx,
		"SELECT * FROM tenant_a.test_analytics_export_control($1,$2,$3)",
		runningExport, runningJob, time.Now().UTC()).Scan(&state, &changedAt)
	if err != nil || state != "cancellation_requested" {
		t.Fatalf("running cancel state=%s err=%v", state, err)
	}
	setAnalyticsDashboardExportLease(
		t, fixture, runningJob, time.Now().UTC().Add(-time.Second),
	)
	var failure []byte
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        tenant_a.test_analytics_export_expired_finalize(
            $1,$2,$3,1,$4,NULL,NULL,$5)`,
		runningExport, runningJob, attemptID, runningLease, time.Now().UTC(),
	).Scan(&state, &changedAt, &failure)
	reserved, _ := analyticsDashboardExportQuota(t, fixture)
	if err != nil || state != "cancelled" || reserved != 0 {
		t.Fatalf("cancel final state=%s reserved=%d err=%v", state, reserved, err)
	}
}

func assertAnalyticsExportDirectMutationDenied(
	t *testing.T, fixture analyticsDashboardExportPG18,
	exportID, jobID, attemptID, leaseID uuid.UUID,
) {
	t.Helper()
	var checkpoint time.Time
	var cancelled bool
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        platform.platform_async_analytics_dashboard_export_checkpoint_v1(
            $1,$2,$3,$4,1,$5,$6)`, asyncTenantA, exportID, jobID, attemptID,
		leaseID, time.Now().UTC()).Scan(&checkpoint, &cancelled)
	if analyticsExportSQLState(err) != "42501" {
		t.Fatalf("direct checkpoint SQLSTATE=%s err=%v", analyticsExportSQLState(err), err)
	}
	var started time.Time
	err = fixture.runtime.QueryRow(fixture.ctx, `SELECT
        platform.platform_async_analytics_dashboard_export_effect_start_v1(
            $1,$2,$3,$4,1,$5,$6,$7,$8)`, asyncTenantA, exportID, jobID,
		attemptID, leaseID, analyticsExportV7(t), analyticsExportDigest(1),
		time.Now().UTC()).Scan(&started)
	if analyticsExportSQLState(err) != "42501" {
		t.Fatalf("direct effect SQLSTATE=%s err=%v", analyticsExportSQLState(err), err)
	}
}

func assertAnalyticsExportExpiredCandidate(
	t *testing.T, fixture analyticsDashboardExportPG18,
	exportID, jobID, attemptID, leaseID, expectedEffectID uuid.UUID,
	expectedRequest []byte,
) {
	t.Helper()
	var tenant, gotExport, gotJob, gotAttempt, gotLease uuid.UUID
	var schema string
	var generation int64
	var gotEffectStarted bool
	var effect pgtype.UUID
	var request []byte
	var cancelled bool
	var expiredAt time.Time
	err := fixture.runtime.QueryRow(fixture.ctx, `SELECT * FROM
        platform.platform_async_analytics_dashboard_export_expired_candidate_v1()`).Scan(
		&tenant, &schema, &gotExport, &gotJob, &gotAttempt, &generation, &gotLease,
		&gotEffectStarted, &effect, &request, &cancelled, &expiredAt)
	expectedEffectStarted := expectedEffectID != uuid.Nil
	if err != nil || tenant != uuid.MustParse(asyncTenantA) || schema != "tenant_a" ||
		gotExport != exportID || gotJob != jobID || gotAttempt != attemptID ||
		gotLease != leaseID || generation != 1 ||
		gotEffectStarted != expectedEffectStarted {
		t.Fatalf("expired candidate tenant=%s export=%s job=%s err=%v",
			tenant, gotExport, gotJob, err)
	}
	if (!expectedEffectStarted && (effect.Valid || len(request) != 0)) ||
		(expectedEffectStarted && (!effect.Valid || uuid.UUID(effect.Bytes) != expectedEffectID ||
			!bytes.Equal(request, expectedRequest))) {
		t.Fatalf("expired candidate effect valid=%v id=%s request=%x",
			effect.Valid, uuid.UUID(effect.Bytes), request)
	}
}
