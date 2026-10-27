//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func rejectIAMTenantReplica(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
) error {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	if _, err := transaction.ExecContext(
		ctx, "SET LOCAL session_replication_role = replica",
	); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("enable %s IAM replica mode: %v", schema, err)
	}
	_, err := transaction.ExecContext(ctx, statement)
	_ = transaction.Rollback()
	if err == nil {
		t.Fatalf("IAM replica-mode statement unexpectedly succeeded in %s", schema)
	}
	return err
}
