package baseline

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
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
		execBaselineTestSQL(t, ctx, database, `
			CREATE TABLE catalog_packages (id integer);
			CREATE FUNCTION catalog_test_function() RETURNS integer
			LANGUAGE sql AS 'SELECT 1';
		`)
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
		execBaselineTestSQL(t, ctx, database, tenantCleanupFixtureSQL)
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
		execBaselineTestSQL(t, ctx, database,
			"CREATE TABLE catalog_manual_data (id integer); INSERT INTO catalog_manual_data VALUES (1);")
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
		execBaselineTestSQL(t, ctx, database, `
			CREATE TABLE catalog_manual_data (id integer);
			CREATE VIEW generic_view AS SELECT id FROM catalog_manual_data;
		`)
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
