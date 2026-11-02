package baseline

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

const baselinePG18TestDSNKey = "HARPY_BASELINE_PG18_TEST_DSN"

func TestCatalogRemovalPostgres18(t *testing.T) {
	dsn := os.Getenv(baselinePG18TestDSNKey)
	if dsn == "" {
		t.Skipf("%s is not set", baselinePG18TestDSNKey)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	t.Run("platform cleanup removes empty module objects", func(t *testing.T) {
		const schema = "platform"
		resetBaselineTestSchema(t, ctx, dsn, schema)
		database, err := open(dsn, schema+", pg_catalog")
		if err != nil {
			t.Fatal(err)
		}
		defer database.Close()
		if _, err = database.ExecContext(ctx, `
			CREATE TABLE catalog_packages (id integer);
			CREATE FUNCTION catalog_test_function() RETURNS integer
			LANGUAGE sql AS 'SELECT 1';
		`); err != nil {
			t.Fatal(err)
		}
		if _, err = apply(ctx, dsn, schema+", pg_catalog",
			platformCatalogRemovalFile, platformCatalogRemovalScript); err != nil {
			t.Fatalf("apply platform Catalog cleanup: %v", err)
		}
		assertNoCatalogObjects(t, ctx, database, schema)
	})

	t.Run("tenant cleanup retires IAM history and removes module objects", func(t *testing.T) {
		const schema = "tenant_catalog_cleanup_test"
		resetBaselineTestSchema(t, ctx, dsn, schema)
		database, err := open(dsn, schema+", pg_catalog")
		if err != nil {
			t.Fatal(err)
		}
		defer database.Close()
		if _, err = database.ExecContext(ctx, tenantCleanupFixtureSQL); err != nil {
			t.Fatal(err)
		}
		script := strings.ReplaceAll(tenantCatalogRemovalScript, tenantSchemaToken, schema)
		if _, err = apply(ctx, dsn, schema+", pg_catalog", tenantCatalogRemovalFile, script); err != nil {
			t.Fatalf("apply tenant Catalog cleanup: %v", err)
		}
		assertNoCatalogObjects(t, ctx, database, schema)

		var retiredResources, retiredPermissions, revokedBindings int
		if err = database.QueryRowContext(ctx,
			"SELECT count(*) FROM iam_resources WHERE state = 'retired'",
		).Scan(&retiredResources); err != nil {
			t.Fatal(err)
		}
		if err = database.QueryRowContext(ctx,
			"SELECT count(*) FROM iam_permissions WHERE state = 'retired'",
		).Scan(&retiredPermissions); err != nil {
			t.Fatal(err)
		}
		for _, table := range []string{
			"iam_role_permission_bindings",
			"iam_user_permission_bindings",
			"iam_organization_permission_bindings",
		} {
			var count int
			if err = database.QueryRowContext(ctx,
				"SELECT count(*) FROM "+table+" WHERE state = 'revoked'",
			).Scan(&count); err != nil {
				t.Fatal(err)
			}
			revokedBindings += count
		}
		if retiredResources != 2 || retiredPermissions != 1 || revokedBindings != 3 {
			t.Fatalf("unexpected IAM cleanup: resources=%d permissions=%d bindings=%d",
				retiredResources, retiredPermissions, revokedBindings)
		}
	})

	t.Run("non-empty Catalog data blocks cleanup", func(t *testing.T) {
		const schema = "tenant_catalog_data_guard_test"
		resetBaselineTestSchema(t, ctx, dsn, schema)
		database, err := open(dsn, schema+", pg_catalog")
		if err != nil {
			t.Fatal(err)
		}
		defer database.Close()
		if _, err = database.ExecContext(ctx,
			"CREATE TABLE catalog_manual_data (id integer); INSERT INTO catalog_manual_data VALUES (1)",
		); err != nil {
			t.Fatal(err)
		}
		script := strings.ReplaceAll(tenantCatalogRemovalScript, tenantSchemaToken, schema)
		if _, err = apply(ctx, dsn, schema+", pg_catalog", tenantCatalogRemovalFile, script); err == nil ||
			!strings.Contains(err.Error(), "explicit data disposition") {
			t.Fatalf("expected fail-closed data disposition error, got %v", err)
		}
		var rows int
		if err = database.QueryRowContext(ctx, "SELECT count(*) FROM catalog_manual_data").Scan(&rows); err != nil {
			t.Fatal(err)
		}
		if rows != 1 {
			t.Fatalf("Catalog data changed after rejected cleanup: rows=%d", rows)
		}
	})

	t.Run("external views block cleanup", func(t *testing.T) {
		const schema = "tenant_catalog_dependency_guard_test"
		resetBaselineTestSchema(t, ctx, dsn, schema)
		database, err := open(dsn, schema+", pg_catalog")
		if err != nil {
			t.Fatal(err)
		}
		defer database.Close()
		if _, err = database.ExecContext(ctx, `
			CREATE TABLE catalog_manual_data (id integer);
			CREATE VIEW generic_view AS SELECT id FROM catalog_manual_data;
		`); err != nil {
			t.Fatal(err)
		}
		script := strings.ReplaceAll(tenantCatalogRemovalScript, tenantSchemaToken, schema)
		if _, err = apply(ctx, dsn, schema+", pg_catalog", tenantCatalogRemovalFile, script); err == nil ||
			!strings.Contains(err.Error(), "view owned outside") {
			t.Fatalf("expected external dependency error, got %v", err)
		}
		var exists bool
		if err = database.QueryRowContext(ctx,
			"SELECT to_regclass($1) IS NOT NULL", schema+".generic_view",
		).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatal("external view was removed by rejected cleanup")
		}
	})
}

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
	defer admin.Close()
	if _, err = admin.ExecContext(ctx,
		"DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE",
	); err != nil {
		t.Fatal(err)
	}
	if err = ensureSchema(ctx, dsn, schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if _, cleanupErr := admin.ExecContext(cleanupCtx,
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
