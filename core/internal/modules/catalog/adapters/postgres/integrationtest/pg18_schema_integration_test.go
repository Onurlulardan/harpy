//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertCatalogMigrationShape(
	t *testing.T, ctx context.Context, database catalogPG18,
) {
	t.Helper()
	for _, tenant := range catalogTenants {
		var tables, functions int
		err := database.owner.QueryRowContext(ctx, `SELECT
 (SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname=$1
  AND tablename LIKE 'catalog_%'),
 (SELECT count(*) FROM pg_catalog.pg_proc AS routine
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=routine.pronamespace
  WHERE namespace.nspname=$1 AND routine.proname LIKE 'catalog_%')`, tenant.schema).
			Scan(&tables, &functions)
		if err != nil || tables != 31 || functions != 38 {
			t.Fatalf("%s Catalog tables/functions=%d/%d err=%v", tenant.schema,
				tables, functions, err)
		}
	}
	assertCatalogRegistryImmutable(t, ctx, database.owner)
	assertCatalogPackageIdentityConstraints(t, ctx, database.owner)
}

func assertCatalogRegistryImmutable(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `UPDATE platform.catalog_packages
SET slug=slug WHERE package_id=$1`, catalogPackageID); err == nil {
		t.Fatal("Catalog registry mutation was accepted")
	}
}
