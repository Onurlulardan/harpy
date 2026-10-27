//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func assertSaaSCustomerGuards(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	controlTenantID string,
) {
	t.Helper()
	insertTestTenant(t, ctx, database, secondTenantID)
	insertTestTenant(t, ctx, database, thirdTenantID)
	insertTestTenant(t, ctx, database, fourthTenantID)
	if err := insertTenantBinding(
		ctx, database, secondTenantID, "saas_customer",
	); err == nil || !strings.Contains(err.Error(), "control binding is required") {
		t.Fatalf("customer-before-control error = %v", err)
	}
	mustInsertTenantBinding(t, ctx, database, controlTenantID, "saas_control")
	mustInsertTenantBinding(t, ctx, database, secondTenantID, "saas_customer")
	mustInsertTenantBinding(t, ctx, database, thirdTenantID, "saas_customer")
	if err := insertTenantBinding(
		ctx, database, fourthTenantID, "saas_control",
	); err == nil || !strings.Contains(err.Error(), "control binding already exists") {
		t.Fatalf("second control error = %v", err)
	}
	assertSaaSCustomerList(t, ctx, database, controlTenantID)
	assertSaaSCustomerTarget(t, ctx, database, secondTenantID, true)
	assertSaaSCustomerTarget(t, ctx, database, controlTenantID, false)
	assertSaaSCustomerOffboard(t, ctx, database, thirdTenantID, true)
	assertSaaSCustomerOffboard(t, ctx, database, controlTenantID, false)
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.tenancy_tenant_bindings(environment_id,tenant_id,purpose)
VALUES ($1,$2,'saas_customer')`, testEnvironmentID, fourthTenantID)
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("runtime direct binding insert error = %v", err)
	}
}

func assertSaaSCustomerList(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	controlTenantID string,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	rows, err := transaction.QueryContext(ctx, `
SELECT tenant_id FROM platform.tenancy_list_saas_customer_tenants_v1($1)`,
		testEnvironmentID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var tenantID string
		if err := rows.Scan(&tenantID); err != nil {
			t.Fatal(err)
		}
		seen[tenantID] = true
	}
	if err := rows.Err(); err != nil || len(seen) != 2 || seen[controlTenantID] ||
		!seen[secondTenantID] || !seen[thirdTenantID] {
		t.Fatalf("SaaS customer list = %#v, %v", seen, err)
	}
}

func assertSaaSCustomerTarget(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID string,
	allowed bool,
) {
	t.Helper()
	assertRuntimeTenantGuard(t, ctx, database,
		"platform.tenancy_assert_saas_customer_target_v1", tenantID, allowed)
}

func assertSaaSCustomerOffboard(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID string,
	allowed bool,
) {
	t.Helper()
	assertRuntimeTenantGuard(t, ctx, database,
		"platform.tenancy_assert_saas_customer_offboard_v1", tenantID, allowed)
}
