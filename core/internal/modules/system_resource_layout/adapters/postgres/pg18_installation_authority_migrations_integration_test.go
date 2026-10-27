//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func applyLayoutMigrationSetPG18(
	t *testing.T, ctx context.Context, database *sql.DB,
	origin, schema string, names []string,
) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	for _, name := range names {
		transaction, err := connection.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		var bound string
		err = transaction.QueryRowContext(ctx,
			"SELECT pg_catalog.set_config('search_path', $1, true)",
			schema+", pg_catalog").Scan(&bound)
		if err == nil {
			_, err = transaction.ExecContext(ctx,
				string(readLayoutMigration(t, origin, name)))
		}
		if err != nil {
			_ = transaction.Rollback()
			t.Fatalf("apply %s to %s: %v", name, schema, err)
		}
		if err := transaction.Commit(); err != nil {
			t.Fatal(err)
		}
	}
}
