//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func TestMeteredResourceOwnerFailureRollsBackAllBillingStateOnPG18(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := billingTestContext(t)
			defer cancel()
			fixture := prepareMeteredResourceFixture(t, ctx, origin)
			statements := loadMeteredAdmissionSQL(t)
			connection := billingRuntimeConnection(t, ctx, fixture.database)
			operationID, ownerFactID := billingV7(t), billingV7(t)
			transaction, err := connection.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			allowed, err := writeMeteredAdmission(
				ctx, transaction, fixture, statements, operationID, ownerFactID,
				"owner-rollback-1",
			)
			if err != nil || !allowed {
				_ = transaction.Rollback()
				t.Fatalf("write Billing admission allowed=%v err=%v", allowed, err)
			}
			// This is the Analytics/API owner write failure in the same transaction.
			if err = transaction.Rollback(); err != nil {
				t.Fatal(err)
			}
			assertMeteredResourceState(
				t, ctx, fixture.database, fixture, operationID, ownerFactID,
				0, 0, 0, 0, 0,
			)

			committedOperation, committedOwner := billingV7(t), billingV7(t)
			transaction, err = connection.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			allowed, err = writeMeteredAdmission(
				ctx, transaction, fixture, statements, committedOperation,
				committedOwner, "owner-commit-1",
			)
			if err != nil || !allowed {
				_ = transaction.Rollback()
				t.Fatalf("commit Billing admission allowed=%v err=%v", allowed, err)
			}
			if err = transaction.Commit(); err != nil {
				t.Fatal(err)
			}
			assertMeteredResourceState(
				t, ctx, fixture.database, fixture, committedOperation,
				committedOwner, 1, 1, 1, 1, 1,
			)
		})
	}
}

func assertMeteredResourceState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture meteredResourceFixture,
	operationID, ownerFactID uuid.UUID,
	wantReceipt, wantProjection, wantDecision, wantEntry int,
	wantGauge int64,
) {
	t.Helper()
	var receipt, projection, decision, entry int
	var gauge int64
	err := database.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.billing_metered_resource_receipts
  WHERE operation_id=$1),
 (SELECT count(*) FROM platform.billing_metered_resources
  WHERE tenant_id=$2 AND resource_kind='api_endpoint' AND owner_fact_id=$3),
 (SELECT count(*) FROM platform.billing_entitlement_decisions
  WHERE tenant_id=$2 AND source_operation_id=$1),
 (SELECT count(*) FROM platform.billing_usage_entries
  WHERE tenant_id=$2 AND owner_fact_id=$1),
 (SELECT gauge_total FROM platform.billing_usage_totals
  WHERE tenant_id=$2 AND meter_period_id=$4)
`, operationID, fixture.tenantID, ownerFactID, fixture.periodID).Scan(
		&receipt, &projection, &decision, &entry, &gauge,
	)
	if err != nil {
		t.Fatal(err)
	}
	if receipt != wantReceipt || projection != wantProjection ||
		decision != wantDecision || entry != wantEntry || gauge != wantGauge {
		t.Fatalf("state receipt=%d projection=%d decision=%d entry=%d gauge=%d",
			receipt, projection, decision, entry, gauge)
	}
}
