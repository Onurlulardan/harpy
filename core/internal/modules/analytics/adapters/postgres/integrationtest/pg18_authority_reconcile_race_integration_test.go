//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type analyticsReconcileResult struct {
	reconciled int
	remaining  int
	err        error
}

func TestAnalyticsAuthorityReconcileConcurrentActivationPostgreSQL184(
	t *testing.T,
) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "authority-race-"+origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))

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

			reconcileConnection, err := database.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_, _ = reconcileConnection.ExecContext(context.Background(),
					`RESET ROLE; RESET SESSION AUTHORIZATION`)
				_ = reconcileConnection.Close()
			}()
			reconcileTransaction, err := reconcileConnection.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer reconcileTransaction.Rollback()
			configureAnalyticsReconcileTransaction(
				t, ctx, reconcileTransaction, true,
			)
			var backendPID int
			if err = reconcileTransaction.QueryRowContext(
				ctx, `SELECT pg_catalog.pg_backend_pid()`,
			).Scan(&backendPID); err != nil {
				t.Fatal(err)
			}

			result := make(chan analyticsReconcileResult, 1)
			go func() {
				var value analyticsReconcileResult
				value.err = reconcileTransaction.QueryRowContext(ctx, `
SELECT reconciled_count,remaining_count
FROM platform.analytics_reconcile_current_tenant_authorities_v2()`,
				).Scan(&value.reconciled, &value.remaining)
				result <- value
			}()
			waitForAnalyticsReconcileLock(t, ctx, database, backendPID)

			if _, err = transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner`); err != nil {
				t.Fatal(err)
			}
			var concurrentAuthorityID string
			if err = transaction.QueryRowContext(ctx, `
SELECT platform.analytics_register_tenant_authority_v1(
    pg_catalog.uuidv7(),$1,1,1,pg_catalog.transaction_timestamp()
)`, analyticsTenantA).Scan(&concurrentAuthorityID); err != nil {
				t.Fatal(err)
			}
			if err = transaction.Commit(); err != nil {
				t.Fatal(err)
			}
			if _, err = blocker.ExecContext(ctx,
				`RESET ROLE; RESET SESSION AUTHORIZATION`); err != nil {
				t.Fatal(err)
			}

			select {
			case value := <-result:
				if value.err != nil {
					t.Fatalf("race-safe Analytics reconcile: %v", value.err)
				}
				if value.reconciled != 1 || value.remaining != 0 {
					t.Fatalf("race reconcile = %d/%d, want 1/0",
						value.reconciled, value.remaining)
				}
				if _, err = reconcileTransaction.ExecContext(ctx,
					"RESET ROLE; RESET SESSION AUTHORIZATION"); err != nil {
					t.Fatal(err)
				}
				if err = reconcileTransaction.Commit(); err != nil {
					t.Fatal(err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("race-safe Analytics reconcile did not resume")
			}
			reconciled, remaining := reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 0 || remaining != 0 {
				t.Fatalf("race replay = %d/%d, want 0/0",
					reconciled, remaining)
			}
		})
	}
}

func configureAnalyticsRuntimeConnection(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(ctx,
		`SELECT pg_catalog.set_config('application_name',$1,false)`,
		"harpy-v2:core-operator-initialize"); err != nil {
		t.Fatal(err)
	}
}

func waitForAnalyticsReconcileLock(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	backendPID int,
) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waitType string
		err := database.QueryRowContext(ctx, `
SELECT COALESCE(wait_event_type,'')
FROM pg_catalog.pg_stat_activity
WHERE pid=$1`, backendPID).Scan(&waitType)
		if err != nil {
			t.Fatal(err)
		}
		if waitType == "Lock" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("Analytics reconcile did not reach the concurrent insert lock")
}
