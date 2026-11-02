package baseline

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

const tenantCleanupFixtureSQL = `
CREATE TABLE iam_resources (
    id uuid PRIMARY KEY, resource_key text NOT NULL, state text NOT NULL,
    retired_at timestamptz, updated_at timestamptz NOT NULL,
    revision bigint NOT NULL, created_at timestamptz NOT NULL
);
CREATE TABLE iam_permissions (
    id uuid PRIMARY KEY, resource_id uuid NOT NULL, state text NOT NULL,
    retired_at timestamptz, updated_at timestamptz NOT NULL,
    revision bigint NOT NULL, created_at timestamptz NOT NULL
);
CREATE TABLE iam_role_permission_bindings (
    permission_id uuid NOT NULL, state text NOT NULL, revoked_at timestamptz,
    updated_at timestamptz NOT NULL, revision bigint NOT NULL
);
CREATE TABLE iam_user_permission_bindings (
    permission_id uuid NOT NULL, state text NOT NULL, revoked_at timestamptz,
    updated_at timestamptz NOT NULL, revision bigint NOT NULL
);
CREATE TABLE iam_organization_permission_bindings (
    permission_id uuid NOT NULL, state text NOT NULL, revoked_at timestamptz,
    updated_at timestamptz NOT NULL, revision bigint NOT NULL
);
CREATE TABLE catalog_menu_definitions (
    definition_id uuid PRIMARY KEY, revision bigint NOT NULL, active_version_id uuid NOT NULL
);
CREATE TABLE catalog_menu_versions (
    version_id uuid PRIMARY KEY, definition_id uuid NOT NULL,
    version_number bigint NOT NULL, content_sha256 text NOT NULL
);
CREATE TABLE catalog_installations (id integer);
CREATE TABLE analytics_catalog_owner_projection (id integer);
CREATE TABLE iam_catalog_component_bindings (id integer);
CREATE FUNCTION catalog_test_function() RETURNS integer LANGUAGE sql AS 'SELECT 1';
CREATE FUNCTION analytics_catalog_test_function() RETURNS integer LANGUAGE sql AS 'SELECT 1';
CREATE FUNCTION iam_catalog_test_function() RETURNS integer LANGUAGE sql AS 'SELECT 1';
INSERT INTO catalog_menu_definitions VALUES (
    '018f6000-0000-7000-8000-000000000001', 1,
    '018f6000-0000-7000-8000-000000000002'
);
INSERT INTO catalog_menu_versions VALUES (
    '018f6000-0000-7000-8000-000000000002',
    '018f6000-0000-7000-8000-000000000001', 1,
    'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
);
INSERT INTO iam_resources VALUES
    ('019f9d00-0000-8000-8000-000000000001', 'catalog', 'active', NULL,
     now() - interval '1 hour', 1, now() - interval '2 hours'),
    ('019f9d00-0000-8000-8000-000000000002', 'catalog.menu', 'active', NULL,
     now() - interval '1 hour', 1, now() - interval '2 hours');
INSERT INTO iam_permissions VALUES (
    '019f9d00-0000-8000-8000-000000000201',
    '019f9d00-0000-8000-8000-000000000001', 'active', NULL,
    now() - interval '1 hour', 1, now() - interval '2 hours'
);
INSERT INTO iam_role_permission_bindings VALUES (
    '019f9d00-0000-8000-8000-000000000201', 'active', NULL,
    now() - interval '1 hour', 1
);
INSERT INTO iam_user_permission_bindings VALUES (
    '019f9d00-0000-8000-8000-000000000201', 'active', NULL,
    now() - interval '1 hour', 1
);
INSERT INTO iam_organization_permission_bindings VALUES (
    '019f9d00-0000-8000-8000-000000000201', 'active', NULL,
    now() - interval '1 hour', 1
);
`

func resetBaselineTestSchema(t *testing.T, ctx context.Context, dsn, schema string) {
	t.Helper()
	admin, err := open(dsn, "pg_catalog")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = admin.ExecContext(ctx,
		"DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE",
	); err != nil {
		_ = admin.Close()
		t.Fatal(err)
	}
	if err = admin.Close(); err != nil {
		t.Fatal(err)
	}
	if err = ensureSchema(ctx, dsn, schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cleanupDB, cleanupErr := open(dsn, "pg_catalog")
		if cleanupErr != nil {
			t.Errorf("open cleanup database for %s: %v", schema, cleanupErr)
			return
		}
		defer cleanupDB.Close()
		if _, cleanupErr = cleanupDB.ExecContext(cleanupCtx,
			"DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE",
		); cleanupErr != nil {
			t.Errorf("drop test schema %s: %v", schema, cleanupErr)
		}
	})
}

func assertNoCatalogObjects(t *testing.T, ctx context.Context, database *sql.DB, schema string) {
	t.Helper()
	var tables, routines int
	if err := database.QueryRowContext(ctx, `
		SELECT count(*)
		FROM pg_catalog.pg_class AS relation
		JOIN pg_catalog.pg_namespace AS namespace
		  ON namespace.oid = relation.relnamespace
		WHERE namespace.nspname = $1
		  AND relation.relkind IN ('r', 'p', 'v', 'm', 'S')
		  AND relation.relname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
	`, schema).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `
		SELECT count(*)
		FROM pg_catalog.pg_proc AS procedure_row
		JOIN pg_catalog.pg_namespace AS namespace
		  ON namespace.oid = procedure_row.pronamespace
		WHERE namespace.nspname = $1
		  AND procedure_row.proname ~ '^(catalog_|analytics_catalog_|iam_catalog_)'
	`, schema).Scan(&routines); err != nil {
		t.Fatal(err)
	}
	if tables != 0 || routines != 0 {
		t.Fatalf("Catalog objects remain in %s: tables=%d routines=%d", schema, tables, routines)
	}
}

func execBaselineTestSQL(t *testing.T, ctx context.Context, database *sql.DB, script string) {
	t.Helper()
	for _, statement := range strings.Split(script, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("execute baseline fixture SQL %q: %v", statement, err)
		}
	}
}
