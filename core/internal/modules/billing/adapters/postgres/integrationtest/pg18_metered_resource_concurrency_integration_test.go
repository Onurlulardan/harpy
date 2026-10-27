//go:build integration

package integrationtest

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type meteredAdmissionOutcome struct {
	allowed bool
	err     error
}

func TestMeteredResourceFiniteBoundaryIsAtomicOnPG18(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	fixture := prepareMeteredResourceFixture(t, ctx, "fresh")
	statements := loadMeteredAdmissionSQL(t)
	connections := []*sql.Conn{
		billingRuntimeConnection(t, ctx, fixture.database),
		billingRuntimeConnection(t, ctx, fixture.database),
	}
	start := make(chan struct{})
	outcomes := make(chan meteredAdmissionOutcome, 2)
	for index, connection := range connections {
		operationID, ownerFactID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
		idempotencyKey := fmt.Sprintf("concurrent-create-%d", index+1)
		go func(
			connection *sql.Conn,
			operationID, ownerFactID uuid.UUID,
			idempotencyKey string,
		) {
			<-start
			transaction, err := connection.BeginTx(ctx, nil)
			if err != nil {
				outcomes <- meteredAdmissionOutcome{err: err}
				return
			}
			allowed, err := writeMeteredAdmission(
				ctx, transaction, fixture, statements, operationID, ownerFactID,
				idempotencyKey,
			)
			if err == nil {
				err = transaction.Commit()
			} else {
				_ = transaction.Rollback()
			}
			outcomes <- meteredAdmissionOutcome{allowed: allowed, err: err}
		}(connection, operationID, ownerFactID, idempotencyKey)
	}
	close(start)
	allowedCount := 0
	for range 2 {
		outcome := <-outcomes
		if outcome.err != nil {
			t.Fatal(outcome.err)
		}
		if outcome.allowed {
			allowedCount++
		}
	}
	if allowedCount != 1 {
		t.Fatalf("finite boundary allowed %d creates, want 1", allowedCount)
	}
	var receipts, projections, decisions, entries int
	var gauge int64
	err := fixture.database.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.billing_metered_resource_receipts
  WHERE tenant_id=$1 AND resource_kind='api_endpoint'),
 (SELECT count(*) FROM platform.billing_metered_resources
  WHERE tenant_id=$1 AND resource_kind='api_endpoint' AND counted_state='active'),
 (SELECT count(*) FROM platform.billing_entitlement_decisions
  WHERE tenant_id=$1 AND entitlement_key='api.active_endpoints'),
 (SELECT count(*) FROM platform.billing_usage_entries
  WHERE tenant_id=$1 AND entitlement_key='api.active_endpoints'
    AND entry_kind='gauge_delta'),
 (SELECT gauge_total FROM platform.billing_usage_totals
  WHERE tenant_id=$1 AND meter_period_id=$2)
`, fixture.tenantID, fixture.periodID).Scan(
		&receipts, &projections, &decisions, &entries, &gauge,
	)
	if err != nil {
		t.Fatal(err)
	}
	if receipts != 2 || projections != 1 || decisions != 2 || entries != 1 || gauge != 1 {
		t.Fatalf("receipts=%d projections=%d decisions=%d entries=%d gauge=%d",
			receipts, projections, decisions, entries, gauge)
	}
}
