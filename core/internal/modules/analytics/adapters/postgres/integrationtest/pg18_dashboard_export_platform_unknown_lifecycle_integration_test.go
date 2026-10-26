//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type analyticsDashboardExportUnknownEvidence struct {
	claim     analyticsPlatformExportClaim
	effect    analyticsPlatformExportEffect
	unknownAt time.Time
}

func assertAnalyticsDashboardExportUnknownLifecycle(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	claim := claimAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsUnknownExportV2, analyticsUnknownJobV2, analyticsUnknownLeaseV2, 1)
	checkpointAt := checkpointAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsUnknownExportV2, analyticsUnknownJobV2, claim, 1,
		analyticsUnknownLeaseV2)
	effect := startAnalyticsDashboardExportPlatformEffect(t, ctx, connection,
		analyticsUnknownExportV2, analyticsUnknownJobV2, analyticsUnknownLeaseV2,
		analyticsUnknownEffectV2, claim, 1, 0xd2, checkpointAt.Add(time.Millisecond))
	evidence := analyticsDashboardExportUnknownEvidence{
		claim: claim, effect: effect, unknownAt: effect.startedAt.Add(time.Millisecond),
	}
	persisted, err := markAnalyticsDashboardExportPlatformUnknown(t, ctx, connection,
		claim.attemptID, analyticsUnknownLeaseV2, effect.effectID, effect.request,
		evidence.unknownAt)
	if err != nil || !persisted {
		t.Fatalf("Platform-first render outcome_unknown = %v: %v", persisted, err)
	}
	assertAnalyticsDashboardExportUnknownEvidence(t, ctx, database, evidence)
	assertAnalyticsExportQuota(t, ctx, database, 0, 2147483648)
	assertAnalyticsDashboardExportUnknownReplay(t, ctx, database, connection, evidence)
	assertAnalyticsDashboardExportUnknownStale(t, ctx, database, connection, evidence)
	assertAnalyticsDashboardExportUnknownDefer(t, ctx, database, connection, evidence)
}

func markAnalyticsDashboardExportPlatformUnknown(t *testing.T, ctx context.Context,
	connection *sql.Conn, attemptID, leaseOwner, effectID string, request []byte,
	observedAt time.Time,
) (bool, error) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var persisted bool
	err := transaction.QueryRowContext(ctx, `SELECT
analytics_dashboard_export_platform_unknown_v1(
't_analytics_a',$1,$2,$3,$4,1,$5,$6,$7,$8)`, analyticsTenantA,
		analyticsUnknownExportV2, analyticsUnknownJobV2, attemptID, leaseOwner,
		effectID, request, observedAt).Scan(&persisted)
	if err != nil {
		return false, err
	}
	if err = transaction.Commit(); err != nil {
		return false, err
	}
	return persisted, nil
}

func assertAnalyticsDashboardExportUnknownReplay(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
	evidence analyticsDashboardExportUnknownEvidence,
) {
	t.Helper()
	before := analyticsDashboardExportUnknownSnapshot(t, ctx, database)
	persisted, err := markAnalyticsDashboardExportPlatformUnknown(t, ctx, connection,
		evidence.claim.attemptID, analyticsUnknownLeaseV2, evidence.effect.effectID,
		evidence.effect.request, evidence.unknownAt)
	if err != nil || !persisted || before !=
		analyticsDashboardExportUnknownSnapshot(t, ctx, database) {
		t.Fatalf("render outcome_unknown replay = %v: %v", persisted, err)
	}
}

func assertAnalyticsDashboardExportUnknownStale(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
	evidence analyticsDashboardExportUnknownEvidence,
) {
	t.Helper()
	for _, testCase := range []struct {
		name, lease, effect, code string
		observed                  time.Time
	}{
		{"owner", analyticsUnknownCompetingV2, evidence.effect.effectID,
			"P6047", evidence.unknownAt},
		{"effect", analyticsUnknownLeaseV2, analyticsUnknownCleanupV2,
			"P6048", evidence.unknownAt},
		{"clock", analyticsUnknownLeaseV2, evidence.effect.effectID,
			"P6048", evidence.unknownAt.Add(time.Microsecond)},
	} {
		before := analyticsDashboardExportUnknownSnapshot(t, ctx, database)
		_, err := markAnalyticsDashboardExportPlatformUnknown(t, ctx, connection,
			evidence.claim.attemptID, testCase.lease, testCase.effect,
			evidence.effect.request, testCase.observed)
		requireAnalyticsPGCode(t, err, testCase.code,
			"render outcome_unknown stale "+testCase.name)
		if before != analyticsDashboardExportUnknownSnapshot(t, ctx, database) {
			t.Fatalf("render outcome_unknown stale %s mutated evidence", testCase.name)
		}
	}
}
