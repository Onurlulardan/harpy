//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

const (
	analyticsMainEffect  = "019b1000-0000-7000-8000-000000000071"
	analyticsMainReceipt = "019b1000-0000-7000-8000-000000000072"
)

type analyticsPlatformExportClaim struct {
	attemptID  string
	claimedAt  time.Time
	leaseUntil time.Time
}

type analyticsPlatformExportEffect struct {
	effectID  string
	request   []byte
	startedAt time.Time
}

func claimAnalyticsDashboardExportPlatform(t *testing.T, ctx context.Context,
	connection *sql.Conn, exportID, jobID, leaseOwner string, generation int64,
) analyticsPlatformExportClaim {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var claim analyticsPlatformExportClaim
	var claimState, disposition, state string
	var storedGeneration int64
	err := transaction.QueryRowContext(ctx, `SELECT claim_state,attempt_id,claimed_at,
lease_until,disposition,state,attempt_generation
FROM analytics_dashboard_export_platform_claim_v1(
't_analytics_a',$1,$2,$3,$4,$5)`, analyticsTenantA, exportID, jobID,
		generation, leaseOwner).Scan(&claimState, &claim.attemptID, &claim.claimedAt,
		&claim.leaseUntil, &disposition, &state, &storedGeneration)
	if err != nil || claimState != "claimed" || disposition != "execute" ||
		state != "running" || storedGeneration != generation ||
		!claim.leaseUntil.After(claim.claimedAt) {
		t.Fatalf("Platform-first claim %s = %s/%s/%s/%d: %v", exportID,
			claimState, disposition, state, storedGeneration, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return claim
}

func checkpointAnalyticsDashboardExportPlatform(t *testing.T, ctx context.Context,
	connection *sql.Conn, exportID, jobID string, claim analyticsPlatformExportClaim,
	generation int64, leaseOwner string,
) time.Time {
	t.Helper()
	observedAt := claim.claimedAt.Add(time.Millisecond)
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var leaseUntil time.Time
	var cancelled bool
	err := transaction.QueryRowContext(ctx, `SELECT lease_until,cancellation_requested
FROM analytics_dashboard_export_platform_checkpoint_v1(
't_analytics_a',$1,$2,$3,$4,$5,$6,$7)`, analyticsTenantA, exportID, jobID,
		claim.attemptID, generation, leaseOwner, observedAt).Scan(&leaseUntil, &cancelled)
	if err != nil || cancelled || !leaseUntil.After(observedAt) {
		t.Fatalf("Platform-first checkpoint %s = %v/%v: %v",
			exportID, leaseUntil, cancelled, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return observedAt
}

func startAnalyticsDashboardExportPlatformEffect(t *testing.T, ctx context.Context,
	connection *sql.Conn, exportID, jobID, leaseOwner, effectID string,
	claim analyticsPlatformExportClaim, generation int64, marker byte,
	startedAt time.Time,
) analyticsPlatformExportEffect {
	t.Helper()
	effect := analyticsPlatformExportEffect{effectID: effectID,
		request: bytes.Repeat([]byte{marker}, 32), startedAt: startedAt}
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var persistedAt time.Time
	err := transaction.QueryRowContext(ctx, `SELECT
analytics_dashboard_export_platform_effect_start_v1(
't_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,$9,'analytics-kek-1',$10,$11,$12,$13,$14,
1024,128,1,'text/csv; charset=utf-8','dashboard.csv',$15)`, analyticsTenantA,
		exportID, jobID, claim.attemptID, generation, leaseOwner, effectID,
		effect.request, "analytics/dashboard/export/"+exportID,
		bytes.Repeat([]byte{marker + 1}, 48), bytes.Repeat([]byte{marker + 2}, 12),
		bytes.Repeat([]byte{marker + 3}, 32), bytes.Repeat([]byte{marker + 4}, 32),
		bytes.Repeat([]byte{marker + 5}, 32), startedAt).Scan(&persistedAt)
	if err != nil || !persistedAt.Equal(startedAt) {
		t.Fatalf("Platform-first effect %s = %v, want %v: %v",
			exportID, persistedAt, startedAt, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	effect.startedAt = persistedAt
	return effect
}

func succeedAnalyticsDashboardExportPlatform(t *testing.T, ctx context.Context,
	connection *sql.Conn, exportID, jobID, leaseOwner, receiptID string,
	claim analyticsPlatformExportClaim, effect analyticsPlatformExportEffect,
	generation int64,
) {
	t.Helper()
	uploadedAt := effect.startedAt.Add(time.Millisecond)
	completedAt := uploadedAt.Add(time.Millisecond)
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var state string
	err := transaction.QueryRowContext(ctx, `SELECT
analytics_dashboard_export_platform_succeed_v1(
't_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'version-2',
'provider-upload-2',$11,$12)`, analyticsTenantA, exportID, jobID,
		claim.attemptID, generation, leaseOwner, effect.effectID, effect.request,
		receiptID, bytes.Repeat([]byte{0xf1}, 32), uploadedAt, completedAt).Scan(&state)
	if err != nil || state != "ready" {
		t.Fatalf("Platform-first succeed %s = %s: %v", exportID, state, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
