//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingTenantQueriesAreIsolatedOnPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	connection := billingRuntimeConnection(t, ctx, database)
	globalAuthorityID := billingGlobalAuthorityID(t, ctx, database)
	a, b := prepareBillingTenantQueries(t, ctx, connection, globalAuthorityID)
	retireTenantQueryOffering(t, ctx, connection, globalAuthorityID, a.offeringVersionID)
	transaction := beginBillingRuntimeSnapshot(t, ctx, connection)
	assertCurrentSubscriptionQuery(t, ctx, transaction, a, a.subscriptionID)
	assertCurrentSubscriptionQuery(t, ctx, transaction, b, b.subscriptionID)
	assertSubscriptionCrossTenantDenied(t, ctx, transaction, a, b)
	assertUsageQuery(t, ctx, transaction, a, 3, 4)
	assertUsageQuery(t, ctx, transaction, b, 8, 11)
	assertUsageCrossTenantDenied(t, ctx, transaction, a, b)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertCurrentSubscriptionQuery(
	t *testing.T,
	ctx context.Context,
	connection billingSnapshotReader,
	fixture billingTenantQueryFixture,
	want uuid.UUID,
) {
	t.Helper()
	authority := resolveTenantQueryAuthority(
		t, ctx, connection, "subscription", fixture.tenantID,
	)
	statement := loadBillingAdapterSQL(t, "subscription", "get_current.sql")
	var count int64
	var got uuid.UUID
	values := make([]any, 23)
	var priceMatches, offeringMatches bool
	var offeringDocument []byte
	arguments := []any{&count, &got}
	for index := range values {
		arguments = append(arguments, &values[index])
	}
	arguments = append(arguments, &priceMatches, &offeringMatches, &offeringDocument)
	if err := connection.QueryRowContext(
		ctx, statement, authority, fixture.tenantID,
	).Scan(arguments...); err != nil || count != 1 || got != want {
		t.Fatalf("current subscription = %s count=%d err=%v, want %s", got, count, err, want)
	}
	var offering struct {
		VersionID uuid.UUID
		State     string
		Public    bool
	}
	if err := json.Unmarshal(offeringDocument, &offering); err != nil ||
		!priceMatches || !offeringMatches || offering.VersionID != fixture.offeringVersionID ||
		offering.State != "retired" || offering.Public {
		t.Fatalf("historical offering matches=%t snapshot=%+v err=%v",
			offeringMatches, offering, err)
	}
}

func assertSubscriptionCrossTenantDenied(
	t *testing.T,
	ctx context.Context,
	connection billingSnapshotReader,
	a, b billingTenantQueryFixture,
) {
	t.Helper()
	statement := loadBillingAdapterSQL(t, "subscription", "get_current.sql")
	err := connection.QueryRowContext(
		ctx, statement, a.authorityID, b.tenantID,
	).Scan(new(int64))
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-tenant subscription error = %v", err)
	}
}

func assertUsageQuery(
	t *testing.T,
	ctx context.Context,
	connection billingSnapshotReader,
	fixture billingTenantQueryFixture,
	wantCounter, wantGauge int64,
) {
	t.Helper()
	authority := resolveTenantQueryAuthority(t, ctx, connection, "usage", fixture.tenantID)
	contextSQL := loadBillingAdapterSQL(t, "usage", "resolve_current_context.sql")
	var baseCount, subscriptionCount int64
	var base, subscription, subscriptionOffering sql.NullString
	if err := connection.QueryRowContext(
		ctx, contextSQL, authority, fixture.tenantID,
	).Scan(&baseCount, &base, &subscriptionCount, &subscription,
		&subscriptionOffering); err != nil || baseCount != 1 || subscriptionCount != 1 {
		t.Fatalf("usage context base=%d subscription=%d err=%v", baseCount, subscriptionCount, err)
	}
	usageSQL := loadBillingAdapterSQL(t, "usage", "get_current.sql")
	rows, err := connection.QueryContext(
		ctx, usageSQL, authority, fixture.tenantID, base.String,
		subscription.String, subscriptionOffering.String,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string]int64{}
	for rows.Next() {
		var periodCount, overrideCount, committed, reserved, gauge int64
		var periodID uuid.UUID
		var key string
		var valueType, meterKind, resetPolicy, enforcement string
		var unlimitedAllowed, unlimited, limitPresent bool
		var periodStart, periodEnd, updatedAt time.Time
		var periodRevision, totalRevision int64
		var limitAmount sql.NullInt64
		if err = rows.Scan(
			&periodCount, &overrideCount, &periodID, &key,
			&valueType, &meterKind, &resetPolicy, &enforcement,
			&unlimitedAllowed, &periodStart, &periodEnd, &periodRevision,
			&committed, &reserved, &gauge, &totalRevision, &updatedAt,
			&limitAmount, &unlimited, &limitPresent,
		); err != nil {
			t.Fatal(err)
		}
		if periodCount != 1 || overrideCount != 0 {
			t.Fatalf("meter counts = %d/%d", periodCount, overrideCount)
		}
		got[key] = committed + reserved + gauge
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	if got["forms.submissions"] != wantCounter || got["documents.storage_bytes"] != wantGauge {
		t.Fatalf("usage totals = %v, want counter=%d gauge=%d", got, wantCounter, wantGauge)
	}
}

func assertUsageCrossTenantDenied(
	t *testing.T,
	ctx context.Context,
	connection billingSnapshotReader,
	a, b billingTenantQueryFixture,
) {
	t.Helper()
	statement := loadBillingAdapterSQL(t, "usage", "get_current.sql")
	rows, err := connection.QueryContext(
		ctx, statement, a.authorityID, b.tenantID, b.offeringVersionID, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("tenant A authority read tenant B usage")
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func resolveTenantQueryAuthority(
	t *testing.T,
	ctx context.Context,
	connection billingSnapshotReader,
	adapter string,
	tenantID uuid.UUID,
) uuid.UUID {
	t.Helper()
	statement := loadBillingAdapterSQL(t, adapter, "resolve_snapshot_authority.sql")
	var authorityID uuid.UUID
	if err := connection.QueryRowContext(ctx, statement, tenantID).Scan(&authorityID); err != nil {
		t.Fatal(err)
	}
	return authorityID
}
