//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func analyticsRuntimeConnection(
	t *testing.T, ctx context.Context, database *sql.DB,
) *sql.Conn {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = connection.ExecContext(ctx,
		`SET SESSION AUTHORIZATION harpy_v2_runtime; SET ROLE harpy_runtime`); err != nil {
		_ = connection.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = connection.ExecContext(context.Background(),
			`RESET ROLE; RESET SESSION AUTHORIZATION`)
		_ = connection.Close()
	})
	return connection
}
