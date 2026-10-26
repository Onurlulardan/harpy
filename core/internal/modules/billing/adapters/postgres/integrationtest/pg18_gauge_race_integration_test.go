//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type gaugeRaceResult struct {
	replay bool
	err    error
}

type gaugeStatements struct {
	lockSnapshot   string
	findEntry      string
	insertDecision string
	insertEntry    string
}

func TestBillingGaugeIdempotencyRaceAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	fixtureConnection := billingRuntimeConnection(t, ctx, database)
	fixture := prepareBillingTenantFixture(
		t, ctx, fixtureConnection, billingGlobalAuthorityID(t, ctx, database),
	)
	statements := gaugeStatements{
		lockSnapshot:   loadBillingAdapterSQL(t, "usage", "lock_snapshot.sql"),
		findEntry:      loadBillingAdapterSQL(t, "usage", "find_gauge_entry.sql"),
		insertDecision: loadBillingAdapterSQL(t, "usage", "insert_decision.sql"),
		insertEntry:    loadBillingAdapterSQL(t, "usage", "insert_gauge_entry.sql"),
	}
	ownerFactID := billingV7(t)
	start := make(chan struct{})
	results := make(chan gaugeRaceResult, 2)
	for range 2 {
		connection := billingRuntimeConnection(t, ctx, database)
		go func() {
			<-start
			results <- adjustGauge(ctx, connection, fixture, ownerFactID, statements)
		}()
	}
	close(start)
	first, second := <-results, <-results
	for _, result := range []gaugeRaceResult{first, second} {
		if result.err != nil {
			t.Fatal(result.err)
		}
	}
	if first.replay == second.replay {
		t.Fatalf("gauge replay flags = %v,%v, want one persisted and one replay", first.replay, second.replay)
	}
	assertGaugeRaceState(t, ctx, database, fixture, ownerFactID)
}

func adjustGauge(
	ctx context.Context,
	connection *sql.Conn,
	fixture billingTenantFixture,
	ownerFactID uuid.UUID,
	statements gaugeStatements,
) gaugeRaceResult {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return gaugeRaceResult{err: err}
	}
	fail := func(err error) gaugeRaceResult {
		_ = transaction.Rollback()
		return gaugeRaceResult{err: err}
	}
	current, limit, err := lockGaugeSnapshot(ctx, transaction, fixture, statements)
	if err != nil {
		return fail(err)
	}
	found, err := findGaugeReplay(ctx, transaction, fixture, ownerFactID, statements.findEntry)
	if err != nil {
		return fail(err)
	}
	if found {
		return gaugeRaceResult{replay: true, err: transaction.Commit()}
	}
	decisionID, entryID := billingV7Must(), billingV7Must()
	var returnedID uuid.UUID
	err = transaction.QueryRowContext(ctx, statements.insertDecision,
		fixture.authorityID, uuid.MustParse(billingTenantID), uuid.New(), decisionID,
		"documents.storage_bytes", fixture.gaugePeriodID, ownerFactID,
		int64(60), current, limit, false, true, false, fixture.now,
	).Scan(&returnedID)
	if err != nil || returnedID != decisionID {
		return fail(fmt.Errorf("insert gauge decision id=%s: %w", returnedID, err))
	}
	err = transaction.QueryRowContext(ctx, statements.insertEntry,
		fixture.authorityID, uuid.MustParse(billingTenantID), entryID, decisionID,
		"documents.storage_bytes", fixture.gaugePeriodID, int64(60), current,
		current+60, fixture.now, ownerFactID,
	).Scan(&returnedID)
	if err != nil || returnedID != entryID {
		return fail(fmt.Errorf("insert gauge entry id=%s: %w", returnedID, err))
	}
	return gaugeRaceResult{err: transaction.Commit()}
}

func lockGaugeSnapshot(
	ctx context.Context,
	transaction *sql.Tx,
	fixture billingTenantFixture,
	statements gaugeStatements,
) (int64, *int64, error) {
	var key, valueKind, meter, reset, enforcement string
	var allowsUnlimited, unlimited bool
	var limit *int64
	var periodID uuid.UUID
	var start, end time.Time
	var committed, reserved, gauge, revision int64
	err := transaction.QueryRowContext(ctx, statements.lockSnapshot,
		fixture.authorityID, uuid.MustParse(billingTenantID),
		"documents.storage_bytes", fixture.gaugePeriodID, "gauge",
	).Scan(&key, &valueKind, &meter, &reset, &enforcement, &allowsUnlimited,
		&limit, &unlimited, &periodID, &start, &end, &committed, &reserved, &gauge,
		&revision)
	return gauge, limit, err
}

func findGaugeReplay(
	ctx context.Context,
	transaction *sql.Tx,
	fixture billingTenantFixture,
	ownerFactID uuid.UUID,
	statement string,
) (bool, error) {
	var entryID, decisionID, periodID, factID, sourceID uuid.UUID
	var key, decisionKey string
	var delta, before, after, current, requested int64
	var createdAt, decidedAt time.Time
	var limit *int64
	var unlimited, allowed, soft bool
	err := transaction.QueryRowContext(ctx, statement, fixture.authorityID,
		uuid.MustParse(billingTenantID), "documents.storage_bytes", ownerFactID,
	).Scan(&entryID, &decisionID, &key, &periodID, &factID, &delta,
		&before, &after, &createdAt, &decisionID, &decisionKey, &periodID,
		&sourceID, &current, &requested, &limit, &unlimited, &allowed, &soft,
		&decidedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func assertGaugeRaceState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture billingTenantFixture,
	ownerFactID uuid.UUID,
) {
	t.Helper()
	var decisions, entries int
	var gauge int64
	err := database.QueryRowContext(ctx, `
SELECT
 (SELECT count(*) FROM platform.billing_entitlement_decisions
  WHERE tenant_id=$1 AND source_operation_id=$2),
 (SELECT count(*) FROM platform.billing_usage_entries
  WHERE tenant_id=$1 AND owner_fact_id=$2), gauge_total
FROM platform.billing_usage_totals WHERE meter_period_id=$3
`, uuid.MustParse(billingTenantID), ownerFactID, fixture.gaugePeriodID).Scan(
		&decisions, &entries, &gauge,
	)
	if err != nil {
		t.Fatal(err)
	}
	if decisions != 1 || entries != 1 || gauge != 60 {
		t.Fatalf("gauge race decisions=%d entries=%d total=%d", decisions, entries, gauge)
	}
}
