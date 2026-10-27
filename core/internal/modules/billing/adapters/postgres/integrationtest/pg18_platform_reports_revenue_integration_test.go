//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertPlatformRevenue(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture platformReportFixture,
) {
	t.Helper()
	assertRevenueRow(
		t, ctx, connection, fixture.a.subscriptionID,
		"monthly", "USD", 1000, 1000, 12000,
	)
	assertRevenueRow(
		t, ctx, connection, fixture.b.subscriptionID,
		"yearly", "EUR", 12018, 1002, 12018,
	)
	assertHalfEvenRevenue(t, ctx, connection)
	assertRevenueTriggerRollback(t, ctx, connection, fixture.a.subscriptionID)
	assertRevenueRebuild(t, ctx, connection, fixture)
}

func assertRevenueRow(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	subscriptionID any,
	wantCadence, wantCurrency string,
	wantSource, wantMonthly, wantAnnual int64,
) {
	t.Helper()
	var cadence, currency, state string
	var source, monthly, annual int64
	err := connection.QueryRowContext(ctx, `
SELECT cadence,currency,state,source_minor_units,
       monthly_minor_units,annual_minor_units
FROM platform.billing_revenue_projection WHERE subscription_id=$1
`, subscriptionID).Scan(&cadence, &currency, &state, &source, &monthly, &annual)
	if err != nil || cadence != wantCadence || currency != wantCurrency ||
		state != "past_due" || source != wantSource || monthly != wantMonthly ||
		annual != wantAnnual {
		t.Fatalf("revenue row=%s/%s/%s %d/%d/%d err=%v",
			cadence, currency, state, source, monthly, annual, err)
	}
}

func assertHalfEvenRevenue(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	for source, want := range map[int64]int64{12006: 1000, 12018: 1002} {
		var got int64
		if err := connection.QueryRowContext(ctx, `
SELECT platform.billing_revenue_monthly_v1($1,'yearly')
`, source).Scan(&got); err != nil || got != want {
			t.Fatalf("half-even %d = %d err=%v, want %d", source, got, err, want)
		}
	}
}

func assertRevenueTriggerRollback(
	t *testing.T, ctx context.Context, connection *sql.Conn, subscriptionID any,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, `
UPDATE platform.billing_subscriptions
SET state='cancelled',grace_ends_at=NULL,revision=revision+1,
    updated_at=pg_catalog.clock_timestamp()
WHERE id=$1`, subscriptionID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = transaction.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_revenue_projection
WHERE subscription_id=$1`, subscriptionID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("projection survived cancelled transaction: %d err=%v", count, err)
	}
	if err = transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err = connection.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_revenue_projection
WHERE subscription_id=$1`, subscriptionID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("projection rollback count=%d err=%v", count, err)
	}
}

func assertRevenueRebuild(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture platformReportFixture,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	first := runRevenueRebuild(t, ctx, connection, fixture)
	if first.sourceCount != 2 || first.projectedCount != 2 ||
		!bytes.Equal(first.sourceSHA, first.projectionSHA) {
		t.Fatalf("first revenue rebuild = %+v", first)
	}
	if _, err := connection.ExecContext(ctx, `
UPDATE platform.billing_revenue_projection SET monthly_minor_units=0
`); billingPGCode(err) != "42501" {
		t.Fatalf("runtime projection DML code=%q err=%v", billingPGCode(err), err)
	}
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(ctx, `
UPDATE platform.billing_revenue_projection
SET monthly_minor_units=monthly_minor_units+1
WHERE subscription_id=$1`, fixture.a.subscriptionID); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	second := runRevenueRebuild(t, ctx, connection, fixture)
	if !bytes.Equal(second.sourceSHA, second.projectionSHA) {
		t.Fatal("revenue rebuild did not repair projection drift")
	}
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	var drifted bool
	if err := connection.QueryRowContext(ctx, `
SELECT before_sha256<>source_sha256 FROM platform.billing_revenue_rebuilds
WHERE id=$1`, second.id).Scan(&drifted); err != nil || !drifted {
		t.Fatalf("revenue drift audit=%t err=%v", drifted, err)
	}
}

type revenueRebuildResult struct {
	id                          any
	sourceCount, projectedCount int64
	sourceSHA, projectionSHA    []byte
}

func runRevenueRebuild(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	fixture platformReportFixture,
) revenueRebuildResult {
	t.Helper()
	result := revenueRebuildResult{id: billingV7(t)}
	now := time.Now().UTC().Truncate(time.Microsecond)
	var ceiling int64
	err := connection.QueryRowContext(ctx, `
SELECT * FROM platform.billing_rebuild_revenue_projection_v1($1,$2,$3,$4)
`, result.id, fixture.actorID, fixture.globalAuthorityID, now).Scan(
		&result.sourceCount, &result.projectedCount, &ceiling,
		&result.sourceSHA, &result.projectionSHA,
	)
	if err != nil {
		t.Fatalf("rebuild revenue projection: %v", err)
	}
	return result
}
