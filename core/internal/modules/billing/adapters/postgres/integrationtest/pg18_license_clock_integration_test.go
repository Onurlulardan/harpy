//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertBillingLicenseClock(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
	); err != nil {
		t.Fatalf("set license clock runtime: %v", err)
	}
	if _, err = connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatalf("set license clock role: %v", err)
	}
	var databaseNow time.Time
	err = connection.QueryRowContext(ctx, `SELECT database_now
FROM platform.billing_license_clock_read_v1()`).Scan(&databaseNow)
	if err != nil || databaseNow.IsZero() {
		t.Fatalf("license database clock=%s err=%v", databaseNow, err)
	}
	if err = database.QueryRowContext(ctx, `SELECT database_now
FROM platform.billing_license_clock_read_v1()`).Scan(&databaseNow); billingPGCode(err) != "42501" {
		t.Fatalf("owner session license clock code=%q err=%v", billingPGCode(err), err)
	}
}
