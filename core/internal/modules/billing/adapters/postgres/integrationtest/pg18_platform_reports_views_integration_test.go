//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func assertPlatformReportViews(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture platformReportFixture,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	assertReportViewCount(t, ctx, connection, "billing_subscription_report_query_v1", 2)
	assertReportViewCount(t, ctx, connection, "billing_audit_report_query_v1", 2)
	assertReportViewCount(t, ctx, connection, "billing_revenue_report_query_v1", 2)
	assertRevenueGroups(t, ctx, connection)
	assertRevenueViewRedacted(t, ctx, connection)
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	assertTransitionTarget(t, ctx, connection, fixture.b, true)
}

func assertReportViewCount(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	view string,
	want int,
) {
	t.Helper()
	var statement string
	switch view {
	case "billing_subscription_report_query_v1":
		statement = "SELECT count(*) FROM platform.billing_subscription_report_query_v1"
	case "billing_audit_report_query_v1":
		statement = "SELECT count(*) FROM platform.billing_audit_report_query_v1"
	case "billing_revenue_report_query_v1":
		statement = "SELECT count(*) FROM platform.billing_revenue_report_query_v1"
	default:
		t.Fatalf("report view %q is not allowlisted", view)
	}
	var got int
	if err := connection.QueryRowContext(ctx, statement).Scan(&got); err != nil || got != want {
		t.Fatalf("%s count=%d err=%v, want %d", view, got, err, want)
	}
}

func assertRevenueGroups(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	rows, err := connection.QueryContext(ctx, `
SELECT currency,currency_registry_version,
       sum(monthly_minor_units),sum(annual_minor_units)
FROM platform.billing_revenue_report_query_v1
GROUP BY currency,currency_registry_version
ORDER BY currency,currency_registry_version`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string][2]int64{}
	for rows.Next() {
		var currency string
		var registry, monthly, annual int64
		if err = rows.Scan(&currency, &registry, &monthly, &annual); err != nil {
			t.Fatal(err)
		}
		got[currency] = [2]int64{monthly, annual}
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	if got["USD"] != [2]int64{1000, 12000} ||
		got["EUR"] != [2]int64{1002, 12018} || len(got) != 2 {
		t.Fatalf("revenue groups = %v", got)
	}
}

func assertRevenueViewRedacted(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	rows, err := connection.QueryContext(ctx, `
SELECT column_name FROM information_schema.columns
WHERE table_schema='platform'
  AND table_name='billing_revenue_report_query_v1'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var column string
		if err = rows.Scan(&column); err != nil {
			t.Fatal(err)
		}
		if column == "tenant_id" || column == "subscription_id" ||
			column == "actor_id" || column == "provider_account_id" {
			t.Fatalf("revenue view exposes %s", column)
		}
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func assertTransitionTarget(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture billingTenantQueryFixture,
	want bool,
) {
	t.Helper()
	statement := loadBillingAdapterSQL(
		t, "platformreports", "resolve_transition_target.sql",
	)
	var tenantID uuid.UUID
	var provider string
	var group, billing int64
	err := connection.QueryRowContext(
		ctx, statement, fixture.subscriptionID,
	).Scan(&tenantID, &provider, &group, &billing)
	if !want && errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err != nil || !want || tenantID != fixture.tenantID || provider != "manual" ||
		group != 1 || billing != 1 {
		t.Fatalf("transition target tenant=%s provider=%s generations=%d/%d err=%v",
			tenantID, provider, group, billing, err)
	}
}

func assertPlatformReportAuthorityCuts(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture platformReportFixture,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx, `
UPDATE platform.cutover_groups SET generation=2
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='billing'
`, fixture.b.tenantID); err != nil {
		t.Fatal(err)
	}
	assertTransitionTarget(t, ctx, connection, fixture.b, false)
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	assertReportViewCount(t, ctx, connection, "billing_subscription_report_query_v1", 1)
	assertReportViewCount(t, ctx, connection, "billing_audit_report_query_v1", 1)
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(ctx, `
UPDATE platform.cutover_groups SET generation=2
WHERE scope_kind='global' AND group_id='billing'`); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	assertReportViewCount(t, ctx, connection, "billing_subscription_report_query_v1", 0)
	assertReportViewCount(t, ctx, connection, "billing_audit_report_query_v1", 0)
	assertReportViewCount(t, ctx, connection, "billing_revenue_report_query_v1", 0)
}
