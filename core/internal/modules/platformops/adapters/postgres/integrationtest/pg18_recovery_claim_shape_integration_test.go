//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"testing"
)

func exerciseRecoveryClaimShape(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx,
		`SET LOCAL SESSION AUTHORIZATION harpy_backup`); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx,
		`SET LOCAL ROLE harpy_backup_reader`); err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx,
		`SET LOCAL application_name = 'harpy-v2:core-database-backup-executor'`); err != nil {
		t.Fatal(err)
	}
	rows, err := transaction.QueryContext(ctx,
		`SELECT * FROM platform.platformops_claim_next_recovery_set_v3($1,$2)`,
		"31313131-3131-7313-8313-313131313131",
		"32323232-3232-7323-8323-323232323232",
	)
	if err != nil {
		t.Fatalf("claim recovery set v3 result shape: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("empty recovery queue returned a claim")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate recovery set v3 claim: %v", err)
	}
}
