//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

const billingAcceptanceA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const billingAcceptanceB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func assertBillingAuthorityReconcile(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	shadowTenantID := insertBillingShadowTenant(t, ctx, database)
	global, remaining := reconcileBillingGlobal(t, ctx, database, true)
	tenant, tenantRemaining := reconcileBillingTenants(t, ctx, database, true)
	if global != 0 || remaining != 0 || tenant != 1 || tenantRemaining != 0 {
		t.Fatalf("initial Billing reconcile = global %d/%d tenant %d/%d",
			global, remaining, tenant, tenantRemaining)
	}
	assertBillingShadowTenantIgnored(t, ctx, database, shadowTenantID)
	global, remaining = reconcileBillingGlobal(t, ctx, database, true)
	tenant, tenantRemaining = reconcileBillingTenants(t, ctx, database, true)
	if global != 0 || remaining != 0 || tenant != 0 || tenantRemaining != 0 {
		t.Fatalf("replay Billing reconcile = global %d/%d tenant %d/%d",
			global, remaining, tenant, tenantRemaining)
	}
	setBillingAuthorityGeneration(
		t, ctx, database, billingEnvironmentID, "global", 2, billingAcceptanceB,
	)
	setBillingAuthorityGeneration(
		t, ctx, database, billingTenantID, "tenant", 2, billingAcceptanceB,
	)
	global, remaining = reconcileBillingGlobal(t, ctx, database, true)
	tenant, tenantRemaining = reconcileBillingTenants(t, ctx, database, true)
	if global != 1 || remaining != 0 || tenant != 1 || tenantRemaining != 0 {
		t.Fatalf("successor Billing reconcile = global %d/%d tenant %d/%d",
			global, remaining, tenant, tenantRemaining)
	}
	reconcileBillingGlobal(t, ctx, database, false)
	reconcileBillingTenants(t, ctx, database, false)
	callBillingReconcileScope(t, ctx, database, "workflow",
		"billing_reconcile_current_global_authority_v2")
}

func reconcileBillingGlobal(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	authorized bool,
) (int, int) {
	t.Helper()
	return callBillingReconcile(t, ctx, database, authorized,
		"billing_reconcile_current_global_authority_v2")
}

func reconcileBillingTenants(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	authorized bool,
) (int, int) {
	t.Helper()
	return callBillingReconcile(t, ctx, database, authorized,
		"billing_reconcile_current_tenant_authorities_v2")
}

func callBillingReconcile(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	authorized bool,
	routine string,
) (int, int) {
	t.Helper()
	scope := ""
	if authorized {
		scope = "billing"
	}
	return callBillingReconcileScope(t, ctx, database, scope, routine)
}

func callBillingReconcileScope(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	scope string,
	routine string,
) (int, int) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(),
			"RESET ROLE; RESET SESSION AUTHORIZATION")
		_ = connection.Close()
	}()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if scope != "" {
		if err = mintBillingReconcileScope(ctx, transaction, scope); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = transaction.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_owner`); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SELECT pg_catalog.set_config('application_name',$1,true)`,
		"harpy-v2:core-operator-initialize"); err != nil {
		t.Fatal(err)
	}
	var reconciled, remaining int
	query := fmt.Sprintf(
		"SELECT reconciled_count,remaining_count FROM platform.%s()", routine,
	)
	err = transaction.QueryRowContext(ctx, query).Scan(&reconciled, &remaining)
	if scope != "billing" {
		if billingPGCode(err) != "42501" {
			t.Fatalf("%s denial code=%q err=%v", routine, billingPGCode(err), err)
		}
		return 0, 0
	}
	if err != nil {
		t.Fatalf("%s: %v", routine, err)
	}
	if _, err = transaction.ExecContext(ctx,
		"RESET ROLE; RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return reconciled, remaining
}
