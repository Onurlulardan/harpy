//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func finalizeBillingTenantAuthorityRace(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) string {
	t.Helper()
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
	first := openBillingFinalizerConnection(t, ctx, database)
	defer closeBillingFinalizerConnection(first)
	second := openBillingFinalizerConnection(t, ctx, database)
	defer closeBillingFinalizerConnection(second)
	firstPID := billingBackendPID(t, ctx, first)
	secondPID := billingBackendPID(t, ctx, second)
	results := make(chan billingSuccessorResult, 2)
	go runBillingTenantSuccessor(ctx, first, results)
	go runBillingTenantSuccessor(ctx, second, results)
	waitForBillingReconcileLock(t, ctx, database, firstPID)
	waitForBillingReconcileLock(t, ctx, database, secondPID)
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var authorityID string
	for range 2 {
		select {
		case result := <-results:
			if result.err != nil {
				t.Fatalf("tenant successor race: %v", result.err)
			}
			if authorityID != "" && result.authorityID != authorityID {
				t.Fatalf("tenant successor race IDs differ: %s != %s",
					result.authorityID, authorityID)
			}
			authorityID = result.authorityID
		case <-time.After(10 * time.Second):
			t.Fatal("tenant successor race did not resume")
		}
	}
	if authorityID == "" {
		t.Fatal("tenant successor race returned no authority")
	}
	return authorityID
}

func runBillingTenantSuccessor(
	ctx context.Context,
	connection *sql.Conn,
	results chan<- billingSuccessorResult,
) {
	authorityID, err := callBillingSuccessor(
		ctx, connection, billingTenantSuccessorOp, billingTenantSuccessorRun,
		billingFormsSuccessorOp, billingFormsSuccessorRun, billingAcceptanceC,
	)
	results <- billingSuccessorResult{authorityID: authorityID, err: err}
}

func billingBackendPID(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
) int {
	t.Helper()
	var backendPID int
	if err := connection.QueryRowContext(ctx,
		`SELECT pg_catalog.pg_backend_pid()`).Scan(&backendPID); err != nil {
		t.Fatal(err)
	}
	return backendPID
}

func assertBillingSuccessorRows(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	globalID string,
	tenantID string,
) {
	t.Helper()
	var globalVersion, tenantVersion, globalCount, tenantCount int
	err := database.QueryRowContext(ctx, `
SELECT
 pg_catalog.uuid_extract_version($1::uuid),
 pg_catalog.uuid_extract_version($2::uuid),
 (SELECT pg_catalog.count(*) FROM platform.billing_global_authorities
  WHERE scope_id=$3 AND authority_group_generation=3),
 (SELECT pg_catalog.count(*) FROM platform.billing_tenant_authorities
  WHERE tenant_id=$4 AND authority_group_generation=3)`,
		globalID, tenantID, billingEnvironmentID, billingTenantID,
	).Scan(&globalVersion, &tenantVersion, &globalCount, &tenantCount)
	if err != nil {
		t.Fatal(err)
	}
	if globalVersion != 7 || tenantVersion != 7 ||
		globalCount != 1 || tenantCount != 1 {
		t.Fatalf("successor rows version/count = %d/%d/%d/%d, want 7/7/1/1",
			globalVersion, tenantVersion, globalCount, tenantCount)
	}
}
