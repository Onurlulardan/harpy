//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type discountEligibilityProbe struct {
	tenantID, subscriptionID, offeringID uuid.UUID
	candidateCount                       int64
	available, offeringAllowed           bool
	alreadyRedeemed                      bool
}

func TestBillingDiscountEligibilityIsIsolatedOnPG18(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	connection := billingRuntimeConnection(t, ctx, database)
	globalAuthorityID := billingGlobalAuthorityID(t, ctx, database)
	a, b := prepareBillingTenantQueries(t, ctx, connection, globalAuthorityID)
	now := time.Now().UTC().Truncate(time.Microsecond)
	discountVersionID := createLimitedDiscount(
		t, ctx, connection, globalAuthorityID, now,
	)
	transaction := beginBillingRuntimeSnapshot(t, ctx, connection)

	resolveSQL := loadBillingAdapterSQL(t, "discounteligibility", "resolve_authority.sql")
	findSQL := loadBillingAdapterSQL(t, "discounteligibility", "find_eligible.sql")
	aTenantAuthority, aGlobalAuthority := resolveEligibilityAuthorities(
		t, ctx, transaction, resolveSQL, a.tenantID,
	)
	bTenantAuthority, bGlobalAuthority := resolveEligibilityAuthorities(
		t, ctx, transaction, resolveSQL, b.tenantID,
	)
	if aGlobalAuthority != bGlobalAuthority || aGlobalAuthority != globalAuthorityID {
		t.Fatalf("global authorities A=%s B=%s want=%s",
			aGlobalAuthority, bGlobalAuthority, globalAuthorityID)
	}
	aProbe, err := queryDiscountEligibility(
		ctx, transaction, findSQL, aTenantAuthority, a.tenantID, aGlobalAuthority, now,
	)
	assertEligibilityProbe(t, aProbe, a, err)
	bProbe, err := queryDiscountEligibility(
		ctx, transaction, findSQL, bTenantAuthority, b.tenantID, bGlobalAuthority, now,
	)
	assertEligibilityProbe(t, bProbe, b, err)
	_, err = queryDiscountEligibility(
		ctx, transaction, findSQL, aTenantAuthority, b.tenantID, aGlobalAuthority, now,
	)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("tenant A authority read tenant B eligibility: %v", err)
	}
	offeringRows, err := transaction.QueryContext(
		ctx,
		loadBillingAdapterSQL(t, "discounteligibility", "list_offerings.sql"),
		globalAuthorityID, discountVersionID,
	)
	if err != nil {
		t.Fatal(err)
	}
	if offeringRows.Next() {
		t.Fatal("all-offerings discount unexpectedly has explicit offerings")
	}
	if err = offeringRows.Close(); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func resolveEligibilityAuthorities(
	t *testing.T,
	ctx context.Context,
	connection billingSnapshotReader,
	statement string,
	tenantID uuid.UUID,
) (uuid.UUID, uuid.UUID) {
	t.Helper()
	var tenantAuthority, globalAuthority uuid.UUID
	if err := connection.QueryRowContext(ctx, statement, tenantID).Scan(
		&tenantAuthority, &globalAuthority,
	); err != nil {
		t.Fatal(err)
	}
	return tenantAuthority, globalAuthority
}

func queryDiscountEligibility(
	ctx context.Context,
	connection billingSnapshotReader,
	statement string,
	tenantAuthority, tenantID, globalAuthority uuid.UUID,
	now time.Time,
) (discountEligibilityProbe, error) {
	result := discountEligibilityProbe{}
	destinations := make([]any, 30)
	values := make([]any, 30)
	for index := range destinations {
		destinations[index] = &values[index]
	}
	destinations[0], destinations[1] = &result.tenantID, &result.candidateCount
	destinations[6], destinations[7] = &result.subscriptionID, &result.offeringID
	destinations[25] = &result.available
	destinations[28], destinations[29] = &result.offeringAllowed, &result.alreadyRedeemed
	err := connection.QueryRowContext(
		ctx, statement, tenantAuthority, tenantID, "RACE10", nil, now, globalAuthority,
	).Scan(destinations...)
	return result, err
}

func assertEligibilityProbe(
	t *testing.T,
	probe discountEligibilityProbe,
	want billingTenantQueryFixture,
	err error,
) {
	t.Helper()
	if err != nil || probe.tenantID != want.tenantID ||
		probe.subscriptionID != want.subscriptionID ||
		probe.offeringID != want.offeringVersionID || probe.candidateCount != 1 ||
		!probe.available || !probe.offeringAllowed || probe.alreadyRedeemed {
		t.Fatalf("eligibility=%+v error=%v want tenant=%s subscription=%s",
			probe, err, want.tenantID, want.subscriptionID)
	}
}
