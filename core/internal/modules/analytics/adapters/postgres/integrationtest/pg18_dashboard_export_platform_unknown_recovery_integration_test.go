//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type analyticsDashboardExportUnknownCleanupClaim struct {
	state      string
	generation sql.NullInt64
	leaseUntil sql.NullTime
}

func assertAnalyticsDashboardExportUnknownDefer(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
	evidence analyticsDashboardExportUnknownEvidence,
) {
	t.Helper()
	observedAt := time.Now().UTC().Truncate(time.Microsecond)
	claim, err := claimAnalyticsDashboardExportUnknownCleanup(t, ctx, connection,
		evidence, analyticsUnknownCleanupV2, observedAt)
	if err != nil || claim.state != "claimed" || !claim.generation.Valid ||
		claim.generation.Int64 != 1 || !claim.leaseUntil.Valid ||
		!claim.leaseUntil.Time.After(observedAt) {
		t.Fatalf("unknown cleanup claim = %+v: %v", claim, err)
	}
	beforeReplay := analyticsDashboardExportUnknownSnapshot(t, ctx, database)
	replay, err := claimAnalyticsDashboardExportUnknownCleanup(t, ctx, connection,
		evidence, analyticsUnknownCleanupV2, observedAt)
	if err != nil || replay.state != "claimed" || replay.generation != claim.generation ||
		replay.leaseUntil != claim.leaseUntil || beforeReplay !=
		analyticsDashboardExportUnknownSnapshot(t, ctx, database) {
		t.Fatalf("unknown cleanup claim replay = %+v: %v", replay, err)
	}
	assertAnalyticsDashboardExportUnknownBusy(t, ctx, database, connection,
		evidence, observedAt)
	checkpointAt := time.Now().UTC().Truncate(time.Microsecond)
	leaseUntil, cancelled, err := checkpointAnalyticsDashboardExportUnknownCleanup(
		t, ctx, connection, evidence, analyticsUnknownCleanupV2, 1, checkpointAt)
	if err != nil || cancelled || !leaseUntil.After(checkpointAt) {
		t.Fatalf("unknown cleanup checkpoint = %v/%v: %v", leaseUntil, cancelled, err)
	}
	assertAnalyticsDashboardExportUnknownCleanupStale(t, ctx, database, connection,
		evidence, checkpointAt)
	deferredAt := time.Now().UTC().Truncate(time.Microsecond)
	retryAt, err := deferAnalyticsDashboardExportUnknownCleanup(t, ctx, connection,
		evidence, analyticsUnknownCleanupV2, 1, deferredAt)
	if err != nil || !retryAt.After(deferredAt) {
		t.Fatalf("unknown cleanup defer = %v: %v", retryAt, err)
	}
	assertAnalyticsDashboardExportUnknownDeferredEvidence(t, ctx, database,
		evidence, retryAt)
	assertAnalyticsExportQuota(t, ctx, database, 0, 2147483648)
	beforeReplay = analyticsDashboardExportUnknownSnapshot(t, ctx, database)
	replayedRetry, err := deferAnalyticsDashboardExportUnknownCleanup(t, ctx,
		connection, evidence, analyticsUnknownCleanupV2, 1, deferredAt)
	if err != nil || !replayedRetry.Equal(retryAt) || beforeReplay !=
		analyticsDashboardExportUnknownSnapshot(t, ctx, database) {
		t.Fatalf("unknown cleanup defer replay = %v/%v: %v",
			replayedRetry, retryAt, err)
	}
	assertAnalyticsDashboardExportUnknownBusy(t, ctx, database, connection,
		evidence, time.Now().UTC().Truncate(time.Microsecond))
}

func claimAnalyticsDashboardExportUnknownCleanup(t *testing.T, ctx context.Context,
	connection *sql.Conn, evidence analyticsDashboardExportUnknownEvidence,
	cleanupOwner string, observedAt time.Time,
) (analyticsDashboardExportUnknownCleanupClaim, error) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var claim analyticsDashboardExportUnknownCleanupClaim
	err := transaction.QueryRowContext(ctx, `SELECT claim_state,cleanup_generation,
cleanup_lease_until FROM analytics_dashboard_export_unknown_cleanup_claim_v1(
't_analytics_a',$1,$2,$3,$4,1,$5,$6,$7,$8,$9)`, analyticsTenantA,
		analyticsUnknownExportV2, analyticsUnknownJobV2, evidence.claim.attemptID,
		analyticsUnknownLeaseV2, evidence.effect.effectID, evidence.effect.request,
		cleanupOwner, observedAt).Scan(&claim.state, &claim.generation, &claim.leaseUntil)
	if err != nil {
		return claim, err
	}
	err = transaction.Commit()
	return claim, err
}

func checkpointAnalyticsDashboardExportUnknownCleanup(t *testing.T,
	ctx context.Context, connection *sql.Conn,
	evidence analyticsDashboardExportUnknownEvidence, cleanupOwner string,
	generation int64, observedAt time.Time,
) (time.Time, bool, error) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var leaseUntil time.Time
	var cancelled bool
	err := transaction.QueryRowContext(ctx, `SELECT cleanup_lease_until,
cancellation_requested FROM analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
't_analytics_a',$1,$2,$3,$4,1,$5,$6,$7,$8,$9,$10)`, analyticsTenantA,
		analyticsUnknownExportV2, analyticsUnknownJobV2, evidence.claim.attemptID,
		analyticsUnknownLeaseV2, evidence.effect.effectID, evidence.effect.request,
		generation, cleanupOwner, observedAt).Scan(&leaseUntil, &cancelled)
	if err != nil {
		return leaseUntil, cancelled, err
	}
	err = transaction.Commit()
	return leaseUntil, cancelled, err
}

func deferAnalyticsDashboardExportUnknownCleanup(t *testing.T, ctx context.Context,
	connection *sql.Conn, evidence analyticsDashboardExportUnknownEvidence,
	cleanupOwner string, generation int64, observedAt time.Time,
) (time.Time, error) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var retryAt time.Time
	err := transaction.QueryRowContext(ctx, `SELECT
analytics_dashboard_export_unknown_cleanup_defer_v1(
't_analytics_a',$1,$2,$3,$4,1,$5,$6,$7,$8,$9,$10)`, analyticsTenantA,
		analyticsUnknownExportV2, analyticsUnknownJobV2, evidence.claim.attemptID,
		analyticsUnknownLeaseV2, evidence.effect.effectID, evidence.effect.request,
		generation, cleanupOwner, observedAt).Scan(&retryAt)
	if err != nil {
		return retryAt, err
	}
	err = transaction.Commit()
	return retryAt, err
}
