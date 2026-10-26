//go:build integration

package migrator

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertBuiltEntrypointPrivilegeDenials(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
) {
	t.Helper()
	runtimeConfig, err := pgx.ParseConfig(database.runtimeDSN)
	if err != nil {
		t.Fatal(err)
	}
	runtimeConfig.RuntimeParams["role"] = "harpy_runtime"
	runtimeConnection, err := pgx.ConnectConfig(ctx, runtimeConfig)
	if err != nil {
		t.Fatalf("connect built runtime denial session: %v", err)
	}
	defer runtimeConnection.Close(context.Background())
	assertBuiltEntrypointStatementDenied(
		t, ctx, runtimeConnection, "CREATE TABLE platform.runtime_ddl_must_fail(id bigint)",
	)
	assertBuiltEntrypointStatementDenied(
		t, ctx, runtimeConnection, "DELETE FROM platform.migration_runs WHERE false",
	)
	assertBuiltEntrypointStatementDenied(t, ctx, runtimeConnection, "SET ROLE harpy_owner")
	migratorConnection, err := pgx.Connect(ctx, database.migratorDSN)
	if err != nil {
		t.Fatalf("connect built migrator denial session: %v", err)
	}
	defer migratorConnection.Close(context.Background())
	assertBuiltEntrypointStatementDenied(t, ctx, migratorConnection, "SET ROLE harpy_runtime")
	assertBuiltEntrypointStatementDenied(
		t, ctx, migratorConnection, "SET ROLE harpy_extension_admin",
	)
}

func assertBuiltEntrypointStatementDenied(
	t *testing.T,
	ctx context.Context,
	connection *pgx.Conn,
	statement string,
) {
	t.Helper()
	_, err := connection.Exec(ctx, statement)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Fatalf("operation was not denied by exact PostgreSQL ACL: %v", err)
	}
}
