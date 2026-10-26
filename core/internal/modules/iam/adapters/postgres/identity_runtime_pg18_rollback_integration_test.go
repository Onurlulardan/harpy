//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func probeIAMIdentityRollback(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	probeIAMIdentityPlatformRollback(t, ctx, database, origin)
	documents := make([]iamFoundationDocument, 0, len(iamIdentityTenantMigrationNames))
	for _, name := range iamIdentityTenantMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path: name, content: readIAMIdentityMigration(t, origin, name),
		})
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path','t_iam_b, pg_catalog',true)"); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	for _, document := range documents {
		if _, err := transaction.ExecContext(ctx, string(document.content)); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("rollback probe apply %s: %v", document.path, err)
		}
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertIAMIdentityRoutineAbsent(t, ctx, database, "t_iam_b", "iam_identity_user_create_v1")
}

func probeIAMIdentityPlatformRollback(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range iamIdentityPlatformMigrationNames {
		if _, err := transaction.ExecContext(ctx,
			string(readIAMPlatformAuthMigration(t, origin, name))); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("platform rollback probe apply %s: %v", name, err)
		}
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertIAMIdentityRoutineAbsent(t, ctx, database, "platform", "iam_identity_directory_upsert_v1")
}

func assertIAMIdentityRoutineAbsent(
	t *testing.T, ctx context.Context, database *sql.DB, schema, routine string,
) {
	t.Helper()
	var count int
	if err := database.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FROM pg_catalog.pg_proc AS routine
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = routine.pronamespace
WHERE namespace.nspname=$1 AND routine.proname=$2`, schema, routine).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback retained %s.%s", schema, routine)
	}
}
