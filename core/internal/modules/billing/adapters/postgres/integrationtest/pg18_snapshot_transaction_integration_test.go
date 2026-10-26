//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

type billingSnapshotReader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func beginBillingRuntimeSnapshot(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
) *sql.Tx {
	t.Helper()
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	transaction, err := connection.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = transaction.Rollback() })
	return transaction
}
