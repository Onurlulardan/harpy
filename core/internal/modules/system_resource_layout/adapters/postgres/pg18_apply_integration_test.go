//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"slices"
	"testing"
)

var layoutTenantSchemas = []string{"layout_a", "layout_b"}

func layoutMigrationRange(
	t *testing.T, first string, last string,
) ([]string, []string) {
	t.Helper()
	start := slices.Index(layoutMigrationNames, first)
	end := slices.Index(layoutMigrationNames, last)
	if start < 0 || end < start {
		t.Fatalf("layout migration range %q..%q is unavailable", first, last)
	}
	return layoutMigrationNames[:start], layoutMigrationNames[start : end+1]
}

func prepareLayoutPG18(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE ROLE harpy_runtime NOLOGIN;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_adopter LOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS;
GRANT harpy_runtime TO harpy_v2_runtime
WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
CREATE SCHEMA platform;
CREATE FUNCTION platform.tenancy_activation_step_context_assert_v1(
    uuid,text,uuid,uuid,name,text
) RETURNS boolean LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS 'SELECT false';
CREATE FUNCTION platform.tenancy_activation_step_consume_v1(
    uuid,text
) RETURNS boolean LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog
AS 'SELECT false';
REVOKE ALL ON FUNCTION
    platform.tenancy_activation_step_context_assert_v1(
        uuid,text,uuid,uuid,name,text
    ),
    platform.tenancy_activation_step_consume_v1(uuid,text)
FROM PUBLIC, harpy_runtime, harpy_v2_runtime;
CREATE SCHEMA layout_a;
CREATE SCHEMA layout_b;
CREATE SCHEMA layout_contract;
GRANT USAGE ON SCHEMA layout_a TO harpy_runtime;
GRANT USAGE ON SCHEMA layout_b TO harpy_runtime;
GRANT USAGE ON SCHEMA layout_contract TO harpy_runtime;
REVOKE CREATE ON SCHEMA layout_a FROM PUBLIC, harpy_runtime;
REVOKE CREATE ON SCHEMA layout_b FROM PUBLIC, harpy_runtime;`)
	if err != nil {
		t.Fatalf("prepare layout database: %v", err)
	}
}

func applyLayoutMigrations(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	for _, schema := range layoutTenantSchemas {
		connection, err := database.Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range layoutMigrationNames {
			transaction, err := connection.BeginTx(ctx, nil)
			if err != nil {
				_ = connection.Close()
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
				_ = connection.Close()
				t.Fatalf("apply %s to %s: %v", name, schema, err)
			}
			if err := transaction.Commit(); err != nil {
				_ = connection.Close()
				t.Fatal(err)
			}
			if origin == "v1_adopted" && schema == "layout_a" &&
				name == "20260716013527_create_system_resource_layout_resolve_v2" {
				seedLegacyLayoutRevision(t, ctx, database)
			}
		}
		_ = connection.Close()
	}
}
