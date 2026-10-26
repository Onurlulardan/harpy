//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertBillingCatalog(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var tables int
	if err := database.QueryRowContext(ctx, `
SELECT count(*) FROM information_schema.tables
WHERE table_schema = 'platform' AND table_name LIKE 'billing\_%' ESCAPE '\'
`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if tables != 92 {
		t.Fatalf("Billing platform table count = %d, want 92", tables)
	}
	var routines int
	if err := database.QueryRowContext(ctx, `
SELECT count(DISTINCT routine_name) FROM information_schema.routines
WHERE routine_schema = 'platform' AND routine_name LIKE 'billing\_%' ESCAPE '\'
`).Scan(&routines); err != nil {
		t.Fatal(err)
	}
	if routines != 160 {
		t.Fatalf("Billing platform routine count = %d, want 160", routines)
	}
}
