//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

func assertRuntimeTenantGuard(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	routine string,
	tenantID string,
	allowed bool,
) {
	t.Helper()
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	statement := fmt.Sprintf("SELECT %s($1,$2)", routine)
	var result string
	err := transaction.QueryRowContext(
		ctx, statement, testEnvironmentID, tenantID,
	).Scan(&result)
	if allowed && (err != nil || result != tenantID) {
		t.Fatalf("%s result = %q, %v", routine, result, err)
	}
	if !allowed && (err == nil || !strings.Contains(err.Error(), "customer target is required")) {
		t.Fatalf("%s denial = %v", routine, err)
	}
}

func assertOnPremiseCustomerRoutinesDenied(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID string,
) {
	t.Helper()
	assertSaaSCustomerTarget(t, ctx, database, tenantID, false)
	assertSaaSCustomerOffboard(t, ctx, database, tenantID, false)
	transaction := beginRuntimeTransaction(t, ctx, database)
	defer transaction.Rollback()
	var listed string
	err := transaction.QueryRowContext(ctx, `
SELECT tenant_id FROM platform.tenancy_list_saas_customer_tenants_v1($1)`,
		testEnvironmentID).Scan(&listed)
	if err == nil || !strings.Contains(err.Error(), "SaaS deployment policy is required") {
		t.Fatalf("on-premise customer list error = %v", err)
	}
}
