//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

type analyticsPlatformRaceResult struct {
	state    string
	schedule bool
	err      error
}

func TestAnalyticsDashboardExportDeleteWinsPlatformFirstClaimRacePG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := prepareAnalyticsPlatformProtocolPG18(t, ctx)
	requestedAt := time.Now().UTC().Add(time.Second)
	assertAnalyticsLegacyRenderABIDenied(t, ctx, database, requestedAt)
	var beforeRevision int64
	if err := database.QueryRowContext(ctx, `SELECT row_revision
FROM t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`,
		analyticsLockExport).Scan(&beforeRevision); err != nil {
		t.Fatal(err)
	}
	blocker, err := database.Conn(ctx)
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
		`SELECT 1 FROM platform.jobs WHERE id=$1 FOR UPDATE`, analyticsLockJob); err != nil {
		t.Fatal(err)
	}
	deleteConnection := analyticsRuntimeConnection(t, ctx, database)
	claimConnection := analyticsRuntimeConnection(t, ctx, database)
	setAnalyticsRaceApplicationName(t, ctx, deleteConnection, "analytics-delete-winner")
	setAnalyticsRaceApplicationName(t, ctx, claimConnection, "analytics-claim-loser")
	deleteResult := make(chan analyticsPlatformRaceResult, 1)
	claimResult := make(chan error, 1)
	go func() {
		deleteResult <- runAnalyticsDeleteWinner(ctx, deleteConnection, requestedAt)
	}()
	waitAnalyticsRaceLock(t, ctx, database, "analytics-delete-winner")
	go func() { claimResult <- runAnalyticsClaimLoser(ctx, claimConnection) }()
	waitAnalyticsRaceLock(t, ctx, database, "analytics-claim-loser")
	if err = blockingTx.Commit(); err != nil {
		t.Fatal(err)
	}
	winner := <-deleteResult
	if winner.err != nil || winner.state != "cancelled" || winner.schedule {
		t.Fatalf("delete winner = %s/%v: %v", winner.state, winner.schedule, winner.err)
	}
	loserErr := <-claimResult
	if loserErr == nil {
		t.Fatal("claim loser unexpectedly committed")
	}
	requireAnalyticsPGCode(t, loserErr, "P6046", "Platform-first claim loser")
	assertAnalyticsClaimLoserMadeNoMutation(t, ctx, database, beforeRevision)
}

func runAnalyticsDeleteWinner(ctx context.Context, connection *sql.Conn,
	requestedAt time.Time,
) (result analyticsPlatformRaceResult) {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		result.err = err
		return
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET LOCAL search_path=t_analytics_a,platform,pg_catalog;
SET LOCAL lock_timeout='10s'; SET LOCAL statement_timeout='15s'`); err != nil {
		result.err = err
		return
	}
	result.err = transaction.QueryRowContext(ctx, `SELECT effective_state,schedule_delete
FROM analytics_dashboard_export_delete_v1('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8)`,
		analyticsLockOperation, analyticsLockDeleteJob, analyticsDashboardActor,
		"analytics-lock-delete", bytes.Repeat([]byte{0xd1}, 32),
		analyticsDashboard, analyticsLockExport, requestedAt).Scan(&result.state, &result.schedule)
	if result.err == nil {
		result.err = transaction.Commit()
	}
	return
}

func runAnalyticsClaimLoser(ctx context.Context, connection *sql.Conn) error {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET LOCAL search_path=t_analytics_a,platform,pg_catalog;
SET LOCAL lock_timeout='10s'; SET LOCAL statement_timeout='15s'`); err != nil {
		return err
	}
	var state string
	err = transaction.QueryRowContext(ctx, `SELECT claim_state
FROM analytics_dashboard_export_platform_claim_v1(
't_analytics_a',$1,$2,$3,1,$4)`, analyticsTenantA, analyticsLockExport,
		analyticsLockJob, analyticsLockLease).Scan(&state)
	if err != nil {
		return err
	}
	return transaction.Commit()
}

func setAnalyticsRaceApplicationName(t *testing.T, ctx context.Context,
	connection *sql.Conn, name string,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx,
		`SELECT pg_catalog.set_config('application_name',$1,false)`, name); err != nil {
		t.Fatal(err)
	}
}

func waitAnalyticsRaceLock(t *testing.T, ctx context.Context, database *sql.DB,
	applicationName string,
) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waiting bool
		err := database.QueryRowContext(ctx, `SELECT COALESCE(bool_or(
wait_event_type='Lock'),false) FROM pg_catalog.pg_stat_activity
WHERE application_name=$1`, applicationName).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not reach the Platform row lock", applicationName)
}

func assertAnalyticsLegacyRenderABIDenied(t *testing.T, ctx context.Context,
	database *sql.DB, observedAt time.Time,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var state string
	err := transaction.QueryRowContext(ctx, `SELECT disposition
FROM analytics_dashboard_export_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
		analyticsLockExport, analyticsLockJob, analyticsLockLease, observedAt).Scan(&state)
	if err == nil {
		t.Fatal("legacy dashboard export render ABI remains executable")
	}
	requireAnalyticsPGCode(t, err, "42501", "legacy dashboard export render ABI")
}
