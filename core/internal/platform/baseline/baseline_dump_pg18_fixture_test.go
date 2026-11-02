package baseline

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func createDumpRoles(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	database := openDumpAdmin(t, ctx, dsn)
	defer database.Close()
	for _, role := range []string{
		"harpy_adopter", "harpy_backup", "harpy_backup_reader", "harpy_cluster_admin",
		"harpy_cutover_authority_owner", "harpy_cutover_decoder", "harpy_cutover_finalizer",
		"harpy_migrator", "harpy_owner", "harpy_reactivation_observer", "harpy_restore",
		"harpy_runtime", "harpy_v1_runtime", "harpy_v2_runtime",
	} {
		var exists bool
		if err := database.QueryRowContext(ctx,
			"SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = $1)", role,
		).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			continue
		}
		if _, err := database.ExecContext(ctx,
			"CREATE ROLE "+pgx.Identifier{role}.Sanitize()+" NOLOGIN",
		); err != nil {
			t.Fatalf("create role %s: %v", role, err)
		}
	}
}

func resetDumpSchemas(t *testing.T, ctx context.Context, dsn, tenantSchema string) {
	t.Helper()
	database := openDumpAdmin(t, ctx, dsn)
	defer database.Close()
	for _, schema := range []string{tenantSchema, "platform", "cutover_authority", "platform_ext"} {
		if _, err := database.ExecContext(ctx,
			"DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE",
		); err != nil {
			t.Fatalf("drop schema %s: %v", schema, err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cleanupDB, err := open(dsn, "pg_catalog")
		if err != nil {
			t.Errorf("open dump cleanup: %v", err)
			return
		}
		defer cleanupDB.Close()
		for _, schema := range []string{tenantSchema, "platform", "cutover_authority", "platform_ext"} {
			if _, err = cleanupDB.ExecContext(cleanupCtx,
				"DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE",
			); err != nil {
				t.Errorf("drop schema %s: %v", schema, err)
			}
		}
	})
}

func assertDumpCleanup(t *testing.T, ctx context.Context, dsn, tenantSchema string) {
	t.Helper()
	database := openDumpAdmin(t, ctx, dsn)
	defer database.Close()
	assertNoCatalogObjects(t, ctx, database, "platform")
	assertNoCatalogObjects(t, ctx, database, tenantSchema)
	assertGooseVersions(t, ctx, database, "platform", 0, platformVersion, 20261102000000)
	assertGooseVersions(t, ctx, database, tenantSchema, 0, tenantVersion, 20261102000001)
	for _, name := range []string{
		"platform.billing_catalog_operations",
		"platform.tenant_migration_requirement_catalogs",
	} {
		var exists bool
		if err := database.QueryRowContext(ctx,
			"SELECT to_regclass($1) IS NOT NULL", name,
		).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatalf("preserved relation %s is missing", name)
		}
	}
	for _, name := range []string{
		"catalog_acl_sha256_v1",
		"catalog_definition_sha256_v1",
	} {
		var count int
		if err := database.QueryRowContext(ctx, `
			SELECT count(*)
			FROM pg_catalog.pg_proc AS procedure_row
			JOIN pg_catalog.pg_namespace AS namespace
			  ON namespace.oid = procedure_row.pronamespace
			WHERE namespace.nspname = 'cutover_authority'
			  AND procedure_row.proname = $1
		`, name).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("preserved routine %s count = %d", name, count)
		}
	}
}

func assertGooseVersions(
	t *testing.T, ctx context.Context, database *sql.DB, schema string, versions ...int64,
) {
	t.Helper()
	rows, err := database.QueryContext(ctx,
		"SELECT version_id FROM "+pgx.Identifier{schema, "goose_db_version"}.Sanitize()+
			" WHERE is_applied ORDER BY version_id",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []int64
	for rows.Next() {
		var version int64
		if err = rows.Scan(&version); err != nil {
			t.Fatal(err)
		}
		got = append(got, version)
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(versions) {
		t.Fatalf("%s goose versions = %v", schema, got)
	}
	for index := range versions {
		if got[index] != versions[index] {
			t.Fatalf("%s goose versions = %v", schema, got)
		}
	}
}

func openDumpAdmin(t *testing.T, ctx context.Context, dsn string) *sql.DB {
	t.Helper()
	database, err := open(dsn, "pg_catalog")
	if err != nil {
		t.Fatal(err)
	}
	if err = database.PingContext(ctx); err != nil {
		_ = database.Close()
		t.Fatal(err)
	}
	return database
}
