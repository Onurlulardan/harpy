//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type usageRaceResult struct {
	allowed       bool
	reservationID uuid.UUID
	err           error
}

type usageStatements struct {
	lockSnapshot      string
	insertDecision    string
	insertReservation string
}

func TestBillingUsageQuotaRaceAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	fixtureConnection := billingRuntimeConnection(t, ctx, database)
	globalAuthorityID := billingGlobalAuthorityID(t, ctx, database)
	fixture := prepareBillingTenantFixture(
		t, ctx, fixtureConnection, globalAuthorityID,
	)
	statements := usageStatements{
		lockSnapshot:      loadBillingAdapterSQL(t, "usage", "lock_snapshot.sql"),
		insertDecision:    loadBillingAdapterSQL(t, "usage", "insert_decision.sql"),
		insertReservation: loadBillingAdapterSQL(t, "usage", "insert_reservation.sql"),
	}
	connections := []*sql.Conn{
		billingRuntimeConnection(t, ctx, database),
		billingRuntimeConnection(t, ctx, database),
	}
	start := make(chan struct{})
	results := make(chan usageRaceResult, 2)
	for index, connection := range connections {
		go func(index int, connection *sql.Conn) {
			<-start
			results <- reserveCounter(ctx, connection, fixture, statements, index)
		}(index, connection)
	}
	close(start)
	first, second := <-results, <-results
	allowed := make([]usageRaceResult, 0, 1)
	for _, result := range []usageRaceResult{first, second} {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.allowed {
			allowed = append(allowed, result)
		}
	}
	if len(allowed) != 1 {
		t.Fatalf("allowed reservations = %d, want 1", len(allowed))
	}
	assertUsageRaceState(t, ctx, database, fixture)
	completeUsageReservation(t, ctx, database, fixture, allowed[0].reservationID)
}

func reserveCounter(
	ctx context.Context,
	connection *sql.Conn,
	fixture billingTenantFixture,
	statements usageStatements,
	index int,
) usageRaceResult {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return usageRaceResult{err: err}
	}
	fail := func(err error) usageRaceResult {
		_ = transaction.Rollback()
		return usageRaceResult{err: err}
	}
	var key, valueKind, meter, reset, enforcement string
	var allowsUnlimited, unlimited bool
	var limit *int64
	var periodID uuid.UUID
	var start, end time.Time
	var committed, reserved, gauge, revision int64
	err = transaction.QueryRowContext(ctx, statements.lockSnapshot,
		fixture.authorityID, uuid.MustParse(billingTenantID), "iam.active_users",
		fixture.counterPeriodID, "counter").Scan(
		&key, &valueKind, &meter, &reset, &enforcement, &allowsUnlimited,
		&limit, &unlimited, &periodID, &start, &end, &committed, &reserved, &gauge,
		&revision,
	)
	if err != nil {
		return fail(fmt.Errorf("lock usage snapshot: %w", err))
	}
	allowed := unlimited || limit != nil && committed+reserved+4 <= *limit
	decisionID, sourceID := billingV7Must(), billingV7Must()
	var insertedID uuid.UUID
	err = transaction.QueryRowContext(ctx, statements.insertDecision,
		fixture.authorityID, uuid.MustParse(billingTenantID), uuid.New(), decisionID,
		key, periodID, sourceID, int64(4), committed+reserved, limit,
		unlimited, allowed, false, fixture.now.Add(time.Duration(index)*time.Microsecond),
	).Scan(&insertedID)
	if err != nil || insertedID != decisionID {
		return fail(fmt.Errorf("insert usage decision: id=%s err=%w", insertedID, err))
	}
	result := usageRaceResult{allowed: allowed}
	if allowed {
		result.reservationID = billingV7Must()
		err = transaction.QueryRowContext(ctx, statements.insertReservation,
			fixture.authorityID, uuid.MustParse(billingTenantID), result.reservationID,
			decisionID, key, periodID, sourceID, int64(4), fixture.now.Add(time.Hour),
			fixture.now.Add(time.Duration(index)*time.Microsecond),
		).Scan(&insertedID)
		if err != nil || insertedID != result.reservationID {
			return fail(fmt.Errorf("insert usage reservation: id=%s err=%w", insertedID, err))
		}
	}
	result.err = transaction.Commit()
	return result
}
