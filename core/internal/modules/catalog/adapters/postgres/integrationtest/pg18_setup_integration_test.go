//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5"
)

const catalogTenantA = "76000000-0000-4000-8000-000000000071"
const catalogTenantB = "76000000-0000-7000-8000-000000000072"

var catalogTenants = []struct {
	id, slug, schema string
	generation       int64
}{
	{catalogTenantA, "catalog-a", "catalog_a", 7},
	{catalogTenantB, "catalog-b", "catalog_b", 9},
}

func prepareCatalogPG18(
	t *testing.T, ctx context.Context, database catalogPG18, origin string,
) {
	prepareCatalogPG18WithInitialAuthority(
		t, ctx, database, origin, "", "",
	)
}

func prepareCatalogPG18InactiveAdoption(
	t *testing.T, ctx context.Context, database catalogPG18,
) {
	prepareCatalogPG18WithInitialAuthority(
		t, ctx, database, "v1_adopted", "shadow", "v1",
	)
}

func prepareCatalogPG18WithInitialAuthority(t *testing.T, ctx context.Context,
	database catalogPG18, origin, state, writer string,
) {
	t.Helper()
	createCatalogRolesAndFoundationTables(t, ctx, database)
	seedCatalogTenantAuthority(t, ctx, database, origin)
	if state != "" {
		setCatalogInitialAuthority(t, ctx, database.owner, state, writer)
	}
	prepareCatalogFoundation(t, ctx, database.owner, origin)
	applyCatalogRuntimeFoundation(t, ctx, database.owner, origin)
	applyCatalogMigrations(t, ctx, database.owner, origin)
	seedCatalogRegistry(t, ctx, database.owner)
	if origin == "v1_adopted" {
		seedCatalogLegacyInventory(t, ctx, database.owner)
	}
}

func setCatalogInitialAuthority(t *testing.T, ctx context.Context,
	database *sql.DB, state, writer string,
) {
	t.Helper()
	for _, tenant := range catalogTenants {
		if _, err := database.ExecContext(ctx, `UPDATE platform.cutover_groups
SET state=$2,writer=$3 WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='catalog'`,
			tenant.id, state, writer); err != nil {
			t.Fatal(err)
		}
		if _, err := database.ExecContext(ctx, `UPDATE platform.domain_cutovers
SET state=$2,writer=$3 WHERE scope_kind='tenant' AND scope_id=$1
AND domain='catalog'`, tenant.id, state, writer); err != nil {
			t.Fatal(err)
		}
	}
}

func createCatalogRolesAndFoundationTables(
	t *testing.T, ctx context.Context, database catalogPG18,
) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_adopter NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD 'runtime-only';
GRANT harpy_runtime TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
CREATE ROLE harpy_extension_admin NOLOGIN SUPERUSER;
CREATE ROLE legacy_vector_owner NOLOGIN SUPERUSER;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA platform TO harpy_runtime, harpy_adopter;
CREATE TABLE platform.tenants(id uuid PRIMARY KEY, slug text UNIQUE NOT NULL,
 schema_name name UNIQUE NOT NULL, status text NOT NULL, deleted_at timestamptz);
CREATE TABLE platform.cutover_groups(scope_kind text, scope_id uuid, group_id text,
 generation bigint, state text, writer text, PRIMARY KEY(scope_kind,scope_id,group_id));
CREATE TABLE platform.cutover_group_members(scope_kind text, scope_id uuid, group_id text,
 ordinal smallint, domain text, PRIMARY KEY(scope_kind,scope_id,group_id,ordinal),
 UNIQUE(scope_kind,scope_id,group_id,domain));
CREATE TABLE platform.domain_cutovers(scope_kind text, scope_id uuid, domain text,
 generation bigint, state text, writer text, PRIMARY KEY(scope_kind,scope_id,domain));
CREATE SCHEMA catalog_a AUTHORIZATION harpy_owner;
CREATE SCHEMA catalog_b AUTHORIZATION harpy_owner;
GRANT USAGE ON SCHEMA catalog_a, catalog_b TO harpy_runtime, harpy_adopter;
CREATE TABLE catalog_a.tenancy_schema_identity(singleton boolean PRIMARY KEY,
 tenant_id uuid NOT NULL, schema_name text NOT NULL, origin text NOT NULL,
 created_at timestamptz NOT NULL);
CREATE TABLE catalog_b.tenancy_schema_identity
 (LIKE catalog_a.tenancy_schema_identity INCLUDING ALL);
GRANT SELECT ON catalog_a.tenancy_schema_identity,
 catalog_b.tenancy_schema_identity TO harpy_runtime;`
	if _, err := database.owner.ExecContext(ctx, statement); err != nil {
		t.Fatal(err)
	}
	applyCatalogCutoverDependencies(t, ctx, database.owner)
	var databaseName string
	if err := database.owner.QueryRowContext(ctx, "SELECT current_database()").Scan(
		&databaseName,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := database.owner.ExecContext(ctx,
		"REVOKE TEMPORARY ON DATABASE "+pgx.Identifier{databaseName}.Sanitize()+" FROM PUBLIC",
	); err != nil {
		t.Fatal(err)
	}
}
