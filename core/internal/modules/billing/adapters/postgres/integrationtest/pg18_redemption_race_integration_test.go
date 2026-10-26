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

type redemptionRaceResult struct {
	inserted          bool
	sourceOperationID uuid.UUID
	err               error
}

type redemptionStatements struct {
	lockRedemption  string
	lockEligibility string
	create          string
	find            string
}

type redemptionEligibility struct {
	discountVersionID uuid.UUID
	policySHA256      []byte
	globalLimit       *int64
	globalUsed        int64
	tenantLimit       *int64
	tenantUsed        int64
	offeringAllowed   bool
}

func TestBillingDiscountRedemptionRaceAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	fixtureConnection := billingRuntimeConnection(t, ctx, database)
	globalAuthorityID := billingGlobalAuthorityID(t, ctx, database)
	tenant := prepareBillingTenantFixture(t, ctx, fixtureConnection, globalAuthorityID)
	fixture := prepareBillingSubscriptionFixture(
		t, ctx, fixtureConnection, globalAuthorityID, tenant,
	)
	discountVersionID := createLimitedDiscount(
		t, ctx, fixtureConnection, globalAuthorityID, fixture.now,
	)
	statements := redemptionStatements{
		lockRedemption:  loadBillingAdapterSQL(t, "discount", "lock_redemption.sql"),
		lockEligibility: loadBillingAdapterSQL(t, "discount", "lock_eligibility.sql"),
		create:          loadBillingAdapterSQL(t, "discount", "create_redemption.sql"),
		find:            loadBillingAdapterSQL(t, "discount", "find_redemption.sql"),
	}
	start := make(chan struct{})
	results := make(chan redemptionRaceResult, 2)
	for range 2 {
		connection := billingRuntimeConnection(t, ctx, database)
		go func() {
			<-start
			results <- redeemLimitedDiscount(ctx, connection, fixture, statements)
		}()
	}
	close(start)
	first, second := <-results, <-results
	winners := make([]redemptionRaceResult, 0, 1)
	for _, result := range []redemptionRaceResult{first, second} {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.inserted {
			winners = append(winners, result)
		}
	}
	if len(winners) != 1 {
		t.Fatalf("discount redemption winners = %d, want 1", len(winners))
	}
	assertRedemptionRaceState(
		t, ctx, database, fixture, discountVersionID,
		winners[0].sourceOperationID, statements.find,
	)
}

func createLimitedDiscount(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	authorityID uuid.UUID,
	now time.Time,
) uuid.UUID {
	t.Helper()
	var result catalogResult
	err := connection.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_discount_v2(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,
 $14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26
)`, authorityID, billingV7(t), billingV7(t), billingV7(t), uuid.New(),
		"redemption-race", billingDigest("redemption-race"), "RACE10",
		"Race discount", "Redemption race fixture", true,
		"basis_point", nil, nil, nil, 1000, "once", nil, 10,
		int64(1), int64(1), true, int16(0), now.Add(-time.Minute), nil, now,
	).Scan(&result.aggregateID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	if err != nil {
		t.Fatal(err)
	}
	return result.versionID
}

func redeemLimitedDiscount(
	ctx context.Context,
	connection *sql.Conn,
	fixture billingSubscriptionFixture,
	statements redemptionStatements,
) redemptionRaceResult {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return redemptionRaceResult{err: err}
	}
	fail := func(err error) redemptionRaceResult {
		_ = transaction.Rollback()
		return redemptionRaceResult{err: err}
	}
	if _, err = transaction.ExecContext(ctx, statements.lockRedemption, "RACE10"); err != nil {
		return fail(fmt.Errorf("lock discount redemption: %w", err))
	}
	eligibility, err := lockRedemptionEligibility(ctx, transaction, fixture, statements)
	if err != nil {
		return fail(err)
	}
	result := redemptionRaceResult{sourceOperationID: billingV7Must()}
	if !eligibility.offeringAllowed || limitReached(eligibility.globalLimit, eligibility.globalUsed) ||
		limitReached(eligibility.tenantLimit, eligibility.tenantUsed) {
		result.err = transaction.Commit()
		return result
	}
	redemptionID := billingV7Must()
	var returnedID uuid.UUID
	err = transaction.QueryRowContext(ctx, statements.create,
		fixture.authorityID, uuid.MustParse(billingTenantID), redemptionID,
		fixture.subscriptionID, eligibility.discountVersionID,
		fixture.offeringVersionID, fixture.priceVersionID,
		result.sourceOperationID, eligibility.policySHA256, fixture.now,
	).Scan(&returnedID)
	if err != nil || returnedID != redemptionID {
		return fail(fmt.Errorf("create discount redemption id=%s: %w", returnedID, err))
	}
	result.inserted = true
	result.err = transaction.Commit()
	return result
}

func lockRedemptionEligibility(
	ctx context.Context,
	transaction *sql.Tx,
	fixture billingSubscriptionFixture,
	statements redemptionStatements,
) (redemptionEligibility, error) {
	result := redemptionEligibility{}
	var code, currency string
	var offeringID, priceID uuid.UUID
	var validFrom, validUntil time.Time
	err := transaction.QueryRowContext(ctx, statements.lockEligibility,
		fixture.authorityID, uuid.MustParse(billingTenantID), "RACE10",
		fixture.subscriptionID, fixture.offeringVersionID, fixture.priceVersionID,
		fixture.now,
	).Scan(&result.discountVersionID, &code, &offeringID, &priceID, &currency,
		&validFrom, &validUntil, &result.globalLimit, &result.globalUsed,
		&result.tenantLimit, &result.tenantUsed, &result.offeringAllowed,
		&result.policySHA256)
	return result, err
}

func limitReached(limit *int64, used int64) bool {
	return limit != nil && used >= *limit
}
