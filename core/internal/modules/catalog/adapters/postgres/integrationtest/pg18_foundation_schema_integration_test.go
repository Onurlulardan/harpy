//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const catalogExtensionDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func prepareCatalogFoundation(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	const tables = `
CREATE SCHEMA platform_ext AUTHORIZATION harpy_owner;
CREATE TABLE platform.environment_identity(singleton boolean PRIMARY KEY,
 environment_id uuid,database_origin text,database_system_identifier text,
 database_oid bigint,release_id uuid);
CREATE TABLE platform.bootstrap_receipts(kind text,run_id uuid,environment_id uuid,
 release_id uuid,database_origin text,database_system_identifier text,database_oid bigint);
CREATE TABLE platform.extension_registry(name text,version text,schema_name text,
 owner_role text,database_origin text,verified_run_id uuid);
CREATE TABLE platform.migration_runs(run_id uuid PRIMARY KEY,release_id uuid,
 environment_id uuid,database_origin text,state text,migration_manifest_sha256 text,
 command text,extension_manifest_sha256 text);
CREATE TABLE platform.schema_contracts(component text PRIMARY KEY,
 database_generation bigint,minimum_runtime_generation bigint,
 migration_manifest_sha256 text,extension_manifest_sha256 text,status text,
 verified_run_id uuid,revision bigint,verified_at timestamptz);
CREATE TABLE platform.release_generations(migration_run_id uuid PRIMARY KEY,
 release_id uuid,environment_id uuid,database_origin text,database_generation bigint,
 minimum_runtime_generation bigint,runtime_generation bigint,
 minimum_database_generation bigint,migration_manifest_sha256 text,
 extension_manifest_sha256 text,postgres_image_digest text,
 bootstrap_receipt_kind text,bootstrap_receipt_run_id uuid,
 extension_receipt_kind text,extension_receipt_run_id uuid,
 cleanup_receipt_kind text,cleanup_receipt_run_id uuid);`
	if _, err := database.ExecContext(ctx, tables); err != nil {
		t.Fatal(err)
	}
	extensionSchema, extensionOwner := "platform_ext", "harpy_extension_admin"
	if origin == "v1_adopted" {
		extensionSchema, extensionOwner = "public", "legacy_vector_owner"
	}
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, "SET LOCAL ROLE "+extensionOwner); err == nil {
		_, err = transaction.ExecContext(ctx, "CREATE EXTENSION vector WITH SCHEMA "+extensionSchema)
	}
	if err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	seedCatalogFoundation(t, ctx, database, origin, extensionSchema, extensionOwner)
	_, err = database.ExecContext(ctx, `GRANT SELECT ON platform.environment_identity,
platform.bootstrap_receipts,platform.extension_registry,platform.migration_runs,
platform.schema_contracts,platform.release_generations TO harpy_runtime`)
	if err != nil {
		t.Fatal(err)
	}
}
