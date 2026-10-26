//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingStorageReservationCommitsActualBytesOnPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	fixture := prepareBillingTenantFixture(t, ctx,
		billingRuntimeConnection(t, ctx, database),
		billingGlobalAuthorityID(t, ctx, database))
	connection := billingRuntimeConnection(t, ctx, database)
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	periodID := resolveStoragePeriod(t, ctx, transaction, fixture)
	reservationID, decisionID := billingV7(t), billingV7(t)
	sourceID, remainderID := billingV7(t), billingV7(t)
	insertStorageReservation(t, ctx, transaction, fixture, periodID,
		reservationID, decisionID, sourceID)
	var returnedID uuid.UUID
	err = transaction.QueryRowContext(ctx,
		loadBillingAdapterSQL(t, "usage", "transition_reservation.sql"),
		fixture.authorityID, uuid.MustParse(billingTenantID), billingV7(t),
		reservationID, int64(1), "reserved", "committed",
		fixture.now.Add(time.Minute), billingV7(t), int64(22), remainderID,
	).Scan(&returnedID)
	if err != nil || returnedID != reservationID {
		_ = transaction.Rollback()
		t.Fatalf("commit storage reservation id=%s error=%v", returnedID, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertStorageReservation(t, ctx, database, fixture, reservationID)
}

func resolveStoragePeriod(t *testing.T, ctx context.Context, transaction *sql.Tx,
	fixture billingTenantFixture,
) uuid.UUID {
	t.Helper()
	var periodID uuid.UUID
	err := transaction.QueryRowContext(ctx,
		loadBillingAdapterSQL(t, "usage", "resolve_period.sql"),
		fixture.authorityID, uuid.MustParse(billingTenantID),
		"documents.storage_bytes",
	).Scan(&periodID)
	if err != nil || periodID != fixture.gaugePeriodID {
		t.Fatalf("resolved storage period=%s error=%v", periodID, err)
	}
	return periodID
}

func insertStorageReservation(t *testing.T, ctx context.Context, transaction *sql.Tx,
	fixture billingTenantFixture, periodID, reservationID, decisionID, sourceID uuid.UUID,
) {
	t.Helper()
	var returnedID uuid.UUID
	err := transaction.QueryRowContext(ctx,
		loadBillingAdapterSQL(t, "usage", "insert_decision.sql"),
		fixture.authorityID, uuid.MustParse(billingTenantID), uuid.New(), decisionID,
		"documents.storage_bytes", periodID, sourceID, int64(30), int64(0),
		int64(100), false, true, false, fixture.now,
	).Scan(&returnedID)
	if err == nil {
		err = transaction.QueryRowContext(ctx,
			loadBillingAdapterSQL(t, "usage", "insert_reservation.sql"),
			fixture.authorityID, uuid.MustParse(billingTenantID), reservationID,
			decisionID, "documents.storage_bytes", periodID, sourceID, int64(30),
			fixture.now.Add(time.Hour), fixture.now,
		).Scan(&returnedID)
	}
	if err != nil || returnedID != reservationID {
		_ = transaction.Rollback()
		t.Fatalf("insert storage reservation id=%s error=%v", returnedID, err)
	}
}

func assertStorageReservation(t *testing.T, ctx context.Context, database *sql.DB,
	fixture billingTenantFixture, reservationID uuid.UUID,
) {
	t.Helper()
	var gauge, reserved, commit, released int64
	err := database.QueryRowContext(ctx, `
SELECT total.gauge_total,total.reserved_total,
 COALESCE(sum(entry.amount_delta) FILTER (WHERE entry.entry_kind='commit'),0),
 COALESCE(sum(entry.amount_delta) FILTER (WHERE entry.entry_kind='release'),0)
FROM platform.billing_usage_totals AS total
JOIN platform.billing_usage_entries AS entry
  ON entry.meter_period_id=total.meter_period_id
WHERE total.meter_period_id=$1 AND entry.reservation_id=$2
GROUP BY total.gauge_total,total.reserved_total
`, fixture.gaugePeriodID, reservationID).Scan(&gauge, &reserved, &commit, &released)
	if err != nil || gauge != 22 || reserved != 0 || commit != 22 || released != -8 {
		t.Fatalf("gauge=%d reserved=%d commit=%d release=%d error=%v",
			gauge, reserved, commit, released, err)
	}
}
