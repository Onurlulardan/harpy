//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

type bindingAttempt struct {
	tenantID string
	err      error
}

func assertConcurrentOnPremiseBinding(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	firstTenantID string,
) {
	t.Helper()
	if err := insertTenantBinding(
		ctx, database, firstTenantID, "saas_control",
	); err == nil || !strings.Contains(err.Error(), "purpose is invalid") {
		t.Fatalf("on-premise SaaS purpose error = %v", err)
	}
	insertTestTenant(t, ctx, database, secondTenantID)
	start := make(chan struct{})
	results := make(chan bindingAttempt, 2)
	for _, tenantID := range []string{firstTenantID, secondTenantID} {
		go func(candidate string) {
			<-start
			results <- bindingAttempt{
				tenantID: candidate,
				err: insertTenantBinding(
					ctx, database, candidate, "onpremise_primary",
				),
			}
		}(tenantID)
	}
	close(start)
	winner, loser := "", ""
	for range 2 {
		result := <-results
		if result.err == nil {
			if winner != "" {
				t.Fatal("both concurrent primary bindings committed")
			}
			winner = result.tenantID
		} else if !strings.Contains(result.err.Error(), "capacity is exhausted") {
			t.Fatalf("concurrent binding error = %v", result.err)
		} else {
			loser = result.tenantID
		}
	}
	if winner == "" || loser == "" {
		t.Fatalf("concurrent winner=%q loser=%q", winner, loser)
	}
	assertOnPremiseSlotNeverReopens(t, ctx, database, winner, loser)
}

func assertOnPremiseSlotNeverReopens(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	winner string,
	loser string,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
UPDATE platform.tenants
SET status='SUSPENDED',deleted_at=statement_timestamp(),updated_at=statement_timestamp()
WHERE id=$1`, winner)
	if err != nil {
		t.Fatal(err)
	}
	if err = insertTenantBinding(
		ctx, database, loser, "onpremise_primary",
	); err == nil || !strings.Contains(err.Error(), "capacity is exhausted") {
		t.Fatalf("reused on-premise slot error = %v", err)
	}
	for _, statement := range []string{
		"UPDATE platform.tenancy_tenant_bindings SET purpose='saas_control' WHERE tenant_id=$1",
		"DELETE FROM platform.tenancy_tenant_bindings WHERE tenant_id=$1",
	} {
		if _, err = database.ExecContext(ctx, statement, winner); err == nil ||
			!strings.Contains(err.Error(), "append-only") {
			t.Fatalf("primary substitution passed: %q error=%v", statement, err)
		}
	}
	assertOnPremiseCustomerRoutinesDenied(t, ctx, database, winner)
}
