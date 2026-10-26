//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type billingReconcileResult struct {
	reconciled int
	remaining  int
	err        error
}

func assertBillingAuthorityReconcileRace(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	tenantID := billingV7(t)
	insertBillingTenantTopology(
		t, ctx, database, tenantID, 7, billingAcceptanceA,
	)
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
		`LOCK TABLE platform.billing_tenant_authorities IN SHARE MODE`,
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
	configureBillingReconcileConnection(t, ctx, reconcileTransaction)
	var backendPID int
	if err = reconcileTransaction.QueryRowContext(ctx,
		`SELECT pg_catalog.pg_backend_pid()`).Scan(&backendPID); err != nil {
		t.Fatal(err)
	}
	result := make(chan billingReconcileResult, 1)
	go func() {
		var value billingReconcileResult
		value.err = reconcileTransaction.QueryRowContext(ctx, `
SELECT reconciled_count,remaining_count
FROM platform.billing_reconcile_current_tenant_authorities_v2()`,
		).Scan(&value.reconciled, &value.remaining)
		result <- value
	}()
	waitForBillingReconcileLock(t, ctx, database, backendPID)
	if _, err = transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner`); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, `
SELECT platform.billing_register_tenant_authority_v1(
    pg_catalog.uuidv7(),$1,7,7,7,pg_catalog.transaction_timestamp()
)`, tenantID); err != nil {
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
			t.Fatalf("race-safe Billing reconcile: %v", value.err)
		}
		if value.reconciled != 0 || value.remaining != 0 {
			t.Fatalf("race reconcile = %d/%d, want 0/0",
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
		t.Fatal("race-safe Billing reconcile did not resume")
	}
	reconciled, remaining := reconcileBillingTenants(t, ctx, database, true)
	if reconciled != 0 || remaining != 0 {
		t.Fatalf("race replay = %d/%d, want 0/0", reconciled, remaining)
	}
}

func configureBillingReconcileConnection(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
) {
	t.Helper()
	if err := mintBillingReconcileScope(ctx, transaction, "billing"); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner`); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx,
		`SELECT pg_catalog.set_config('application_name',$1,true)`,
		"harpy-v2:core-operator-initialize"); err != nil {
		t.Fatal(err)
	}
}

func waitForBillingReconcileLock(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	backendPID int,
) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waitType string
		if err := database.QueryRowContext(ctx, `
SELECT COALESCE(wait_event_type,'')
FROM pg_catalog.pg_stat_activity
WHERE pid=$1`, backendPID).Scan(&waitType); err != nil {
			t.Fatal(err)
		}
		if waitType == "Lock" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("Billing reconcile did not reach the concurrent insert lock")
}
