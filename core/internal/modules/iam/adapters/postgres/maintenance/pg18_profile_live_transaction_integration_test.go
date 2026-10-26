//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"
)

func beginProfileRuntimeTransaction(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) *sql.Tx {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			t.Fatal(err)
		}
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		schema+", platform, pg_catalog"); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	return transaction
}
