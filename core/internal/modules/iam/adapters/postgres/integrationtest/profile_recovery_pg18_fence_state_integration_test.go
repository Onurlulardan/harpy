//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func assertRecoveryRawFenceState(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	conn := openRecoveryRuntimeConnection(t, ctx, db)
	tx := beginRecoveryRuntimeRole(t, ctx, conn, db)
	defer func() { _ = tx.Rollback(context.Background()) }()
	var contract []byte
	err := tx.QueryRow(ctx, `SELECT iam_profile_write_contract_read_v2(
$1::name,'user'::text)`, db.tenant.schema).Scan(&contract)
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != "P5741" {
		t.Fatalf("raw fenced write contract = %T/%v", err, err)
	}
}
