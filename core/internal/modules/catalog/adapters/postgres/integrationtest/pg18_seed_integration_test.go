//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

const catalogPackageID = "018f2000-0000-7000-8000-000000000001"
const catalogComponentID = "018f2000-0000-7000-8000-000000000002"
const catalogManifestDigest = "1111111111111111111111111111111111111111111111111111111111111111"

func seedCatalogRegistry(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	at := time.Date(2026, 7, 18, 18, 30, 0, 123000, time.UTC)
	_, err := database.ExecContext(ctx, `INSERT INTO platform.catalog_packages
(package_id,slug,created_at) VALUES($1,'catalog-evidence',$2)`, catalogPackageID, at)
	if err == nil {
		_, err = database.ExecContext(ctx, `INSERT INTO platform.catalog_package_versions
(package_id,version,manifest,manifest_sha256,signature,key_id,release_id,registered_at)
VALUES($1,'1.0.0','{}'::jsonb,$3,decode(repeat('22',64),'hex'),'catalog-key-1',
'018f2000-0000-7000-8000-000000000099',$2)`, catalogPackageID, at, catalogManifestDigest)
	}
	if err != nil {
		t.Fatalf("seed Catalog registry: %v", err)
	}
}

func seedCatalogLegacyInventory(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	for _, tenant := range catalogTenants {
		createCatalogLegacyTables(t, ctx, database, tenant.schema)
		version := "1.0.0"
		if tenant.id == catalogTenantB {
			version = "9.9.9"
		}
		_, err := database.ExecContext(ctx, `INSERT INTO `+tenant.schema+`.modules
(id,tenant_id,slug,version,deleted_at) VALUES($1,$2,'catalog-evidence',$3,NULL)`,
			catalogPackageID, tenant.id, version)
		if err == nil {
			_, err = database.ExecContext(ctx, `INSERT INTO `+tenant.schema+`.module_schema_versions
(id,tenant_id,module_slug,module_version,manifest_hash)
VALUES($1,$2,'catalog-evidence',$3,$4)`, catalogLegacySchemaVersionID(tenant.id),
				tenant.id, version, catalogManifestDigest)
		}
		if err != nil {
			t.Fatalf("seed %s legacy Catalog inventory: %v", tenant.schema, err)
		}
	}
}

func createCatalogLegacyTables(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	statement := `CREATE TABLE ` + schema + `.modules(
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,slug text NOT NULL,
 version text NOT NULL,deleted_at timestamptz);
CREATE TABLE ` + schema + `.module_dependencies(
 id uuid PRIMARY KEY,module_id uuid,depends_on_module_id uuid);
CREATE TABLE ` + schema + `.module_forms(
 id uuid PRIMARY KEY,module_id uuid,form_config_id uuid);
CREATE TABLE ` + schema + `.module_schema_versions(
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,module_slug text NOT NULL,
 module_version text NOT NULL,manifest_hash text NOT NULL);
CREATE TABLE ` + schema + `.module_migration_runs(
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL);
CREATE TABLE ` + schema + `.menu_configs(
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,config jsonb NOT NULL,deleted_at timestamptz);
GRANT SELECT ON ` + schema + `.modules,` + schema + `.module_dependencies,
` + schema + `.module_forms,` + schema + `.module_schema_versions,
` + schema + `.module_migration_runs,` + schema + `.menu_configs
TO harpy_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatal(err)
	}
}

func catalogLegacySchemaVersionID(tenantID string) string {
	if tenantID == catalogTenantA {
		return "018f2000-0000-7000-8000-000000000031"
	}
	return "018f2000-0000-7000-8000-000000000032"
}
