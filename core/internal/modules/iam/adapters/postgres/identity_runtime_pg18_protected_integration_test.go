//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertIAMIdentityProtectedAdmin(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a",
		"51000000-0000-4000-8000-000000000002")
	target := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx,
		"SET SESSION AUTHORIZATION harpy_v2_runtime"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := connection.ExecContext(context.Background(),
			"RESET SESSION AUTHORIZATION"); err != nil {
			t.Errorf("reset protected-admin runtime session: %v", err)
		}
	}()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		"pg_catalog, t_iam_a"); err != nil {
		t.Fatal(err)
	}
	row := transaction.QueryRowContext(ctx, `
SELECT result_id,result_revision,authorization_clock,replayed
FROM iam_identity_user_disable_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6,$7,$8)`,
		"a3000000-0000-4000-8000-000000000004", bytes.Repeat([]byte{0x34}, 32),
		iamIdentityActorID, target.revision,
		"51000000-0000-4000-8000-000000000002", iamIdentityPermissionID,
		actor.tokenVersion, actor.authorizationVersion)
	var id string
	var revision, clock int64
	var replayed bool
	queryErr := row.Scan(&id, &revision, &clock, &replayed)
	if queryErr == nil {
		queryErr = transaction.Commit()
	}
	assertIAMIdentitySQLState(t, queryErr, "23514")
}
