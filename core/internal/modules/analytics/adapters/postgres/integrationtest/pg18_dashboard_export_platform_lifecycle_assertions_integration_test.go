//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsDashboardExportPlatformClaimReplay(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn,
	claim analyticsPlatformExportClaim,
) {
	t.Helper()
	before := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExportID, analyticsExportJob)
	replay := claimAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseA, 1)
	after := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExportID, analyticsExportJob)
	if replay.attemptID != claim.attemptID || !replay.claimedAt.Equal(claim.claimedAt) ||
		!replay.leaseUntil.Equal(claim.leaseUntil) || before != after {
		t.Fatal("Platform-first claim replay changed exact evidence")
	}

	transaction := analyticsExportTransaction(t, ctx, connection)
	_, err := transaction.ExecContext(ctx, `SELECT claim_state FROM
analytics_dashboard_export_platform_claim_v1(
't_analytics_a',$1,$2,$3,1,$4)`, analyticsTenantA, analyticsExportID,
		analyticsExportJob, analyticsExportLeaseB)
	_ = transaction.Rollback()
	requireAnalyticsPGCode(t, err, "P6046", "Platform-first stale claim")
	if analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExportID, analyticsExportJob) != after {
		t.Fatal("Platform-first stale claim mutated exact evidence")
	}
}

func assertAnalyticsDashboardExportEffectReplay(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn, claim analyticsPlatformExportClaim,
	effect analyticsPlatformExportEffect,
) {
	t.Helper()
	before := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExportID, analyticsExportJob)
	replay := startAnalyticsDashboardExportPlatformEffect(t, ctx, connection,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseA,
		analyticsMainEffect, claim, 1, 0xb1, effect.startedAt)
	after := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExportID, analyticsExportJob)
	if replay.effectID != effect.effectID || !replay.startedAt.Equal(effect.startedAt) ||
		before != after {
		t.Fatal("Platform-first effect replay changed exact evidence")
	}
}

func assertAnalyticsDashboardExportSuccessReplay(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn, claim analyticsPlatformExportClaim,
	effect analyticsPlatformExportEffect,
) {
	t.Helper()
	before := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExportID, analyticsExportJob)
	succeedAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseA,
		analyticsMainReceipt, claim, effect, 1)
	after := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExportID, analyticsExportJob)
	if before != after {
		t.Fatal("Platform-first success replay changed exact evidence")
	}
}

func assertAnalyticsDashboardExportStoredState(t *testing.T, ctx context.Context,
	database *sql.DB, exportID, state, quota string,
) {
	t.Helper()
	var storedState, storedQuota string
	err := database.QueryRowContext(ctx, `SELECT state,quota_state FROM
t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`, exportID).Scan(
		&storedState, &storedQuota)
	if err != nil || storedState != state || storedQuota != quota {
		t.Fatalf("Analytics export %s state = %s/%s: %v",
			exportID, storedState, storedQuota, err)
	}
}
