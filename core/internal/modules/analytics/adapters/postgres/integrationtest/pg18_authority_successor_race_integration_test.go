//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type analyticsSuccessorResult struct {
	authorityID string
	err         error
}

func TestAnalyticsCutoverSuccessorAuthorityRacePostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "successor-race-"+origin)
			fixture := prepareAnalyticsSuccessorPG18(t, ctx, database, origin)
			seedAnalyticsSuccessorOperation(t, ctx, database,
				analyticsSuccessorOperationB, analyticsSuccessorRunB,
				"analytics", "analytics")

			blocker, err := database.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer blocker.Close()
			transaction, err := blocker.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer transaction.Rollback()
			if _, err = transaction.ExecContext(ctx,
				`LOCK TABLE platform.analytics_tenant_authorities IN SHARE MODE`,
			); err != nil {
				t.Fatal(err)
			}

			first := openAnalyticsFinalizerConnection(t, ctx, database)
			defer closeAnalyticsFinalizerConnection(first)
			second := openAnalyticsFinalizerConnection(t, ctx, database)
			defer closeAnalyticsFinalizerConnection(second)
			firstPID := analyticsBackendPID(t, ctx, first)
			secondPID := analyticsBackendPID(t, ctx, second)
			results := make(chan analyticsSuccessorResult, 2)
			go runAnalyticsSuccessorCall(ctx, first, analyticsSuccessorOperationA,
				analyticsSuccessorRunA, fixture.digest, results)
			go runAnalyticsSuccessorCall(ctx, second, analyticsSuccessorOperationB,
				analyticsSuccessorRunB, fixture.digest, results)
			waitForAnalyticsReconcileLock(t, ctx, database, firstPID)
			waitForAnalyticsReconcileLock(t, ctx, database, secondPID)
			if err = transaction.Commit(); err != nil {
				t.Fatal(err)
			}

			var successful string
			failures := 0
			for range 2 {
				select {
				case result := <-results:
					if result.err == nil {
						if successful != "" {
							t.Fatal("competing successor operations both succeeded")
						}
						successful = result.authorityID
					} else {
						requireAnalyticsPGCode(t, result.err, "P6101",
							"competing successor operation")
						failures++
					}
				case <-time.After(10 * time.Second):
					t.Fatal("successor race did not resume")
				}
			}
			if successful == "" || failures != 1 {
				t.Fatalf("successor race success/failures = %q/%d",
					successful, failures)
			}
			var count int
			if err = database.QueryRowContext(ctx, `SELECT pg_catalog.count(*)
FROM platform.analytics_tenant_authorities
WHERE tenant_id=$1 AND authority_group_generation=2`,
				analyticsTenantA).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("successor race current authorities = %d, want 1", count)
			}
		})
	}
}

func openAnalyticsFinalizerConnection(
	t *testing.T, ctx context.Context, database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	configureAnalyticsFinalizerConnection(t, ctx, connection)
	return connection
}

func closeAnalyticsFinalizerConnection(connection *sql.Conn) {
	_, _ = connection.ExecContext(context.Background(),
		`RESET SESSION AUTHORIZATION`)
	_ = connection.Close()
}

func analyticsBackendPID(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) int {
	t.Helper()
	var backendPID int
	if err := connection.QueryRowContext(ctx,
		`SELECT pg_catalog.pg_backend_pid()`).Scan(&backendPID); err != nil {
		t.Fatal(err)
	}
	return backendPID
}

func runAnalyticsSuccessorCall(
	ctx context.Context, connection *sql.Conn, operationID, runID, digest string,
	results chan<- analyticsSuccessorResult,
) {
	authorityID, err := callAnalyticsSuccessor(
		ctx, connection, operationID, runID, digest,
	)
	results <- analyticsSuccessorResult{authorityID: authorityID, err: err}
}
