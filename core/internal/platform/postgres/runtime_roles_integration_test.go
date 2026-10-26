//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func prepareRuntimeRoles(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	statements := []string{
		"CREATE ROLE harpy_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_forms_record_reader NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_forms_record_writer NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '" + runtimeIntegrationPassword + "'",
		"CREATE ROLE harpy_migrator LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '" + runtimeIntegrationPassword + "'",
		"CREATE ROLE harpy_extension_admin NOLOGIN NOINHERIT SUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"GRANT harpy_runtime TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_forms_record_reader TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_forms_record_writer TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_owner TO harpy_migrator WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"ALTER DATABASE harpy_runtime OWNER TO harpy_owner",
		"REVOKE CONNECT, TEMPORARY ON DATABASE harpy_runtime FROM PUBLIC",
		"GRANT CONNECT ON DATABASE harpy_runtime TO harpy_runtime",
		"GRANT CONNECT ON DATABASE harpy_runtime TO harpy_v2_runtime",
		"GRANT CONNECT ON DATABASE harpy_runtime TO harpy_migrator",
	}
	executeRuntimeSetup(t, ctx, admin, statements)
}

func prepareRuntimeCatalog(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	statements := []string{
		"CREATE SCHEMA platform AUTHORIZATION harpy_owner",
		"CREATE SCHEMA platform_ext AUTHORIZATION harpy_owner",
		"SET ROLE harpy_extension_admin",
		"CREATE EXTENSION vector WITH SCHEMA platform_ext",
		"RESET ROLE",
		"SET ROLE harpy_owner",
		`CREATE TABLE platform.environment_identity (
singleton boolean PRIMARY KEY, environment_id uuid NOT NULL, database_origin text NOT NULL,
database_system_identifier text NOT NULL, database_oid bigint NOT NULL, release_id uuid NOT NULL)`,
		`CREATE TABLE platform.tenants (
id uuid PRIMARY KEY, slug text COLLATE "C" NOT NULL UNIQUE, schema_name text NOT NULL UNIQUE,
status text NOT NULL, deleted_at timestamp without time zone)`,
		`CREATE FUNCTION platform.tenancy_runtime_resolve_command_v1(p_tenant_id uuid)
RETURNS TABLE(schema_name text,slug text,status text,not_deleted boolean)
LANGUAGE sql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
SELECT tenant.schema_name,tenant.slug,tenant.status,tenant.deleted_at IS NULL
FROM platform.tenants AS tenant WHERE tenant.id=p_tenant_id FOR SHARE
$function$`,
		`CREATE TABLE platform.bootstrap_receipts (
kind text NOT NULL, run_id uuid NOT NULL, environment_id uuid NOT NULL,
release_id uuid NOT NULL, database_origin text NOT NULL,
database_system_identifier text NOT NULL, database_oid bigint NOT NULL)`,
		`CREATE TABLE platform.extension_registry (
name text NOT NULL, version text NOT NULL, schema_name text NOT NULL,
owner_role text NOT NULL, database_origin text NOT NULL, verified_run_id uuid NOT NULL)`,
		`CREATE TABLE platform.migration_runs (
run_id uuid PRIMARY KEY, release_id uuid NOT NULL, environment_id uuid NOT NULL,
database_origin text NOT NULL, state text NOT NULL,
command text NOT NULL,
migration_manifest_sha256 text NOT NULL, extension_manifest_sha256 text NOT NULL)`,
		`CREATE TABLE platform.schema_contracts (
component text PRIMARY KEY, database_generation bigint NOT NULL,
minimum_runtime_generation bigint NOT NULL, migration_manifest_sha256 text NOT NULL,
extension_manifest_sha256 text NOT NULL, status text NOT NULL,
verified_run_id uuid NOT NULL, revision bigint NOT NULL, verified_at timestamptz NOT NULL)`,
		`CREATE TABLE platform.database_generation_manifests (
environment_id uuid NOT NULL, database_generation bigint NOT NULL,
migration_manifest_sha256 text NOT NULL, extension_manifest_sha256 text NOT NULL,
PRIMARY KEY(environment_id,database_generation),
UNIQUE(environment_id,database_generation,migration_manifest_sha256,extension_manifest_sha256))`,
		`CREATE TABLE platform.release_generations (
migration_run_id uuid PRIMARY KEY, release_id uuid NOT NULL, environment_id uuid NOT NULL,
database_origin text NOT NULL, database_generation bigint NOT NULL,
minimum_runtime_generation bigint NOT NULL, runtime_generation bigint NOT NULL,
minimum_database_generation bigint NOT NULL, migration_manifest_sha256 text NOT NULL,
extension_manifest_sha256 text NOT NULL, postgres_image_digest text NOT NULL,
bootstrap_receipt_kind text NOT NULL, bootstrap_receipt_run_id uuid NOT NULL,
extension_receipt_kind text NOT NULL, extension_receipt_run_id uuid NOT NULL,
cleanup_receipt_kind text NOT NULL, cleanup_receipt_run_id uuid NOT NULL,
CHECK(minimum_database_generation<=database_generation),
CHECK(minimum_runtime_generation<=runtime_generation),
FOREIGN KEY(environment_id,database_generation,migration_manifest_sha256,extension_manifest_sha256)
REFERENCES platform.database_generation_manifests
(environment_id,database_generation,migration_manifest_sha256,extension_manifest_sha256))`,
		"RESET ROLE",
		"GRANT USAGE ON SCHEMA platform TO harpy_runtime",
		"GRANT SELECT ON ALL TABLES IN SCHEMA platform TO harpy_runtime",
		`REVOKE ALL ON FUNCTION
platform.tenancy_runtime_resolve_command_v1(uuid) FROM PUBLIC`,
		`GRANT EXECUTE ON FUNCTION
platform.tenancy_runtime_resolve_command_v1(uuid) TO harpy_runtime`,
	}
	executeRuntimeSetup(t, ctx, admin, statements)
	seedRuntimeFoundation(t, ctx, admin)
}

func executeRuntimeSetup(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
	statements []string,
) {
	t.Helper()
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare runtime database: %v", err)
		}
	}
}
