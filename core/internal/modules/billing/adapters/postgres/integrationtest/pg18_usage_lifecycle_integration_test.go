//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertUsageRaceState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingTenantFixture,
) {
	t.Helper()
	var decisions, reservations int
	var committed, reserved int64
	if err := database.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.billing_entitlement_decisions
  WHERE tenant_id=$1 AND entitlement_key='forms.submissions'),
 (SELECT count(*) FROM platform.billing_usage_reservations
  WHERE tenant_id=$1 AND entitlement_key='forms.submissions'),
 committed_total, reserved_total
FROM platform.billing_usage_totals WHERE meter_period_id=$2
`, uuid.MustParse(billingTenantID), fixture.counterPeriodID).Scan(
		&decisions, &reservations, &committed, &reserved,
	); err != nil {
		t.Fatal(err)
	}
	if decisions != 2 || reservations != 1 || committed != 0 || reserved != 4 {
		t.Fatalf("usage race decisions=%d reservations=%d committed=%d reserved=%d",
			decisions, reservations, committed, reserved)
	}
}

func completeUsageReservation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingTenantFixture,
	reservationID uuid.UUID,
) {
	t.Helper()
	connection := billingRuntimeConnection(t, ctx, database)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var returnedID uuid.UUID
	err = transaction.QueryRowContext(ctx,
		loadBillingAdapterSQL(t, "usage", "start_effect.sql"),
		fixture.authorityID, uuid.MustParse(billingTenantID), reservationID,
		int64(1), fixture.now.Add(time.Minute), billingV7(t),
	).Scan(&returnedID)
	if err != nil || returnedID != reservationID {
		_ = transaction.Rollback()
		t.Fatalf("start usage effect id=%s err=%v", returnedID, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	transaction, err = connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = transaction.QueryRowContext(ctx,
		loadBillingAdapterSQL(t, "usage", "transition_reservation.sql"),
		fixture.authorityID, uuid.MustParse(billingTenantID), billingV7(t),
		reservationID, int64(2), "reserved", "committed",
		fixture.now.Add(2*time.Minute), billingV7(t), int64(4), uuid.Nil,
	).Scan(&returnedID)
	if err != nil || returnedID != reservationID {
		_ = transaction.Rollback()
		t.Fatalf("complete usage reservation id=%s err=%v", returnedID, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var state string
	var committed, reserved int64
	if err = database.QueryRowContext(ctx, `
SELECT reservation.state,total.committed_total,total.reserved_total
FROM platform.billing_usage_reservations AS reservation
JOIN platform.billing_usage_totals AS total
  ON total.meter_period_id=reservation.meter_period_id
WHERE reservation.id=$1
`, reservationID).Scan(&state, &committed, &reserved); err != nil {
		t.Fatal(err)
	}
	if state != "committed" || committed != 4 || reserved != 0 {
		t.Fatalf("usage completion state=%s committed=%d reserved=%d",
			state, committed, reserved)
	}
}
