//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	billingv1adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/billingv1adoption"
	"github.com/google/uuid"
)

func assertBillingV1Rollback(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	store billingv1adoptionapp.ScopeStore,
	operation billingv1adoptionapp.Operation,
	plan billingv1adoptionapp.ScopePlan,
	now time.Time,
) {
	t.Helper()
	rehearsedAt := now.Add(time.Second)
	if _, err := billingv1adoptionapp.RehearseRollback(
		ctx, store, operation, uuid.New(), strings.Repeat("cd", 32), rehearsedAt,
	); err == nil {
		t.Fatal("changed Billing V1 source was accepted for rollback")
	}
	if _, err := billingv1adoptionapp.RehearseRollback(
		ctx, store, operation, uuid.New(), plan.SourceSHA256, rehearsedAt,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(
		ctx, "SELECT platform.billing_assert_v1_adoption_ready_v1($1,$2,$3)",
		operation.OperationID, operation.RunID, plan.SourceSHA256,
	); err != nil {
		t.Fatal(err)
	}
}

func assertBillingV1OpeningIdentity(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	var evidenceCount, entryCount int
	var fabricatedOperations int
	err := connection.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.billing_v1_adoption_opening_evidence),
 (SELECT count(*) FROM platform.billing_usage_entries WHERE entry_kind='opening'),
 (SELECT count(*) FROM platform.billing_usage_entries
  WHERE entry_kind='opening' AND source_operation_id IS NOT NULL)
`).Scan(&evidenceCount, &entryCount, &fabricatedOperations)
	if err != nil {
		t.Fatal(err)
	}
	if evidenceCount != 4 || entryCount != 4 || fabricatedOperations != 0 {
		t.Fatalf("opening evidence=%d entries=%d fabricated=%d",
			evidenceCount, entryCount, fabricatedOperations)
	}
}
