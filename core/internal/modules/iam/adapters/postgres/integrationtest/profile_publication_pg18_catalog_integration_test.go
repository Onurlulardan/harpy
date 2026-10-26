//go:build integration

package integrationtest

import (
	"context"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

const publicationExtensionDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

const publicationCatalogSQL = `
CREATE ROLE harpy_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_forms_record_reader NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_forms_record_writer NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '` + publicationPassword + `';
CREATE ROLE harpy_migrator LOGIN NOINHERIT NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD '` + publicationPassword + `';
CREATE ROLE harpy_adopter LOGIN NOINHERIT;
CREATE ROLE harpy_cluster_admin NOLOGIN NOINHERIT;
CREATE ROLE harpy_extension_admin NOLOGIN NOINHERIT SUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
GRANT harpy_runtime TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
GRANT harpy_forms_record_reader TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
GRANT harpy_forms_record_writer TO harpy_v2_runtime WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
GRANT harpy_owner TO harpy_migrator WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
CREATE SCHEMA platform AUTHORIZATION harpy_owner;
CREATE SCHEMA platform_ext AUTHORIZATION harpy_owner;
CREATE SCHEMA t_profile_publication AUTHORIZATION harpy_owner;
SET ROLE harpy_extension_admin;
CREATE EXTENSION vector WITH SCHEMA platform_ext;
RESET ROLE;
SET ROLE harpy_owner;
CREATE TABLE platform.environment_identity(
 singleton boolean PRIMARY KEY,environment_id uuid NOT NULL,database_origin text NOT NULL,
 database_system_identifier text NOT NULL,database_oid bigint NOT NULL,release_id uuid NOT NULL);
CREATE TABLE platform.tenants(
 id uuid PRIMARY KEY,slug text COLLATE "C" NOT NULL UNIQUE,schema_name text NOT NULL UNIQUE,
 status text NOT NULL,deleted_at timestamp without time zone);
CREATE TABLE platform.domain_cutovers(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,domain text NOT NULL,state text NOT NULL,
 writer text NOT NULL,generation bigint NOT NULL,PRIMARY KEY(scope_kind,scope_id,domain),
 UNIQUE(scope_kind,scope_id,domain,generation));
CREATE TABLE platform.cutover_groups(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,group_id text NOT NULL,state text NOT NULL,
 writer text NOT NULL,generation bigint NOT NULL,PRIMARY KEY(scope_kind,scope_id,group_id),
 UNIQUE(scope_kind,scope_id,group_id,generation));
CREATE TABLE platform.cutover_group_members(
 scope_kind text NOT NULL,scope_id uuid NOT NULL,group_id text NOT NULL,
 ordinal smallint NOT NULL,domain text NOT NULL,PRIMARY KEY(scope_kind,scope_id,group_id,ordinal),
 FOREIGN KEY(scope_kind,scope_id,group_id) REFERENCES platform.cutover_groups(scope_kind,scope_id,group_id),
 FOREIGN KEY(scope_kind,scope_id,domain) REFERENCES platform.domain_cutovers(scope_kind,scope_id,domain),
 UNIQUE(scope_kind,scope_id,group_id,domain));
CREATE TABLE platform.bootstrap_receipts(
 kind text NOT NULL,run_id uuid NOT NULL,environment_id uuid NOT NULL,release_id uuid NOT NULL,
 database_origin text NOT NULL,database_system_identifier text NOT NULL,database_oid bigint NOT NULL);
CREATE TABLE platform.extension_registry(
 name text NOT NULL,version text NOT NULL,schema_name text NOT NULL,owner_role text NOT NULL,
 database_origin text NOT NULL,verified_run_id uuid NOT NULL);
CREATE TABLE platform.migration_runs(
 run_id uuid PRIMARY KEY,release_id uuid NOT NULL,environment_id uuid NOT NULL,
 database_origin text NOT NULL,state text NOT NULL,command text NOT NULL,
 migration_manifest_sha256 text NOT NULL,extension_manifest_sha256 text NOT NULL);
CREATE TABLE platform.schema_contracts(
 component text PRIMARY KEY,database_generation bigint NOT NULL,minimum_runtime_generation bigint NOT NULL,
 migration_manifest_sha256 text NOT NULL,extension_manifest_sha256 text NOT NULL,status text NOT NULL,
 verified_run_id uuid NOT NULL,revision bigint NOT NULL,verified_at timestamptz NOT NULL);
CREATE TABLE platform.database_generation_manifests(
 environment_id uuid NOT NULL,database_generation bigint NOT NULL,migration_manifest_sha256 text NOT NULL,
 extension_manifest_sha256 text NOT NULL,PRIMARY KEY(environment_id,database_generation),
 UNIQUE(environment_id,database_generation,migration_manifest_sha256,extension_manifest_sha256));
CREATE TABLE platform.release_generations(
 migration_run_id uuid PRIMARY KEY,release_id uuid NOT NULL,environment_id uuid NOT NULL,
 database_origin text NOT NULL,database_generation bigint NOT NULL,minimum_runtime_generation bigint NOT NULL,
 runtime_generation bigint NOT NULL,minimum_database_generation bigint NOT NULL,
 migration_manifest_sha256 text NOT NULL,extension_manifest_sha256 text NOT NULL,
 postgres_image_digest text NOT NULL,bootstrap_receipt_kind text NOT NULL,bootstrap_receipt_run_id uuid NOT NULL,
 extension_receipt_kind text NOT NULL,extension_receipt_run_id uuid NOT NULL,
 cleanup_receipt_kind text NOT NULL,cleanup_receipt_run_id uuid NOT NULL);
INSERT INTO platform.tenants VALUES(
 '10000000-0000-4000-8000-000000000001','profile-publication','t_profile_publication','ACTIVE',NULL);
INSERT INTO platform.domain_cutovers VALUES
 ('tenant','10000000-0000-4000-8000-000000000001','iam','v2','v2',1),
 ('tenant','10000000-0000-4000-8000-000000000001','system_resource_layout','v2','v2',1);
INSERT INTO platform.cutover_groups VALUES(
 'tenant','10000000-0000-4000-8000-000000000001','iam_system_resources','v2','v2',1);
INSERT INTO platform.cutover_group_members VALUES
 ('tenant','10000000-0000-4000-8000-000000000001','iam_system_resources',1,'iam'),
 ('tenant','10000000-0000-4000-8000-000000000001','iam_system_resources',2,'system_resource_layout');
CREATE TABLE t_profile_publication.tenancy_schema_identity(
 singleton boolean PRIMARY KEY CHECK(singleton),tenant_id uuid NOT NULL UNIQUE,
 schema_name text NOT NULL,origin text NOT NULL,created_at timestamptz NOT NULL);
INSERT INTO t_profile_publication.tenancy_schema_identity VALUES(
 true,'10000000-0000-4000-8000-000000000001','t_profile_publication',$1,pg_catalog.now());
RESET ROLE;`

func preparePublicationCatalog(t *testing.T, ctx context.Context, db publicationDatabase) {
	t.Helper()
	tenantOrigin := db.tenant.origin
	if tenantOrigin == "fresh" {
		tenantOrigin = "new"
	}
	if _, err := db.admin.Exec(ctx, publicationCatalogSQL,
		pgx.QueryExecModeSimpleProtocol, tenantOrigin); err != nil {
		t.Fatalf("prepare publication catalog: %v", err)
	}
	database := pgx.Identifier{db.admin.Config().Database}.Sanitize()
	statements := []string{
		"ALTER DATABASE " + database + " OWNER TO harpy_owner",
		"REVOKE CONNECT, TEMPORARY ON DATABASE " + database + " FROM PUBLIC",
		"GRANT CONNECT ON DATABASE " + database +
			" TO harpy_runtime, harpy_v2_runtime, harpy_migrator",
	}
	for _, statement := range statements {
		if _, err := db.admin.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare publication database ACL: %v", err)
		}
	}
	applyPublicationRuntimeAuthorityDocuments(t, ctx, db)
}

func seedPublicationStartup(
	t *testing.T, ctx context.Context, db publicationDatabase, migrationDigest string,
) {
	t.Helper()
	const sql = `SET ROLE harpy_owner;
INSERT INTO platform.environment_identity
SELECT true,$1,'fresh',system_identifier::text,database.oid::bigint,$2
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database
WHERE database.datname=pg_catalog.current_database();
INSERT INTO platform.migration_runs VALUES($3,$2,$1,'fresh','succeeded','up',$4,$5);
INSERT INTO platform.schema_contracts VALUES('core',1,1,$4,$5,'ready',$3,1,pg_catalog.now());
INSERT INTO platform.bootstrap_receipts
SELECT kind,run_id,$1,$2,'fresh',system_identifier::text,database.oid::bigint
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database,
(VALUES('cluster_bootstrap'::text,$6::uuid),('extension_admin',$7::uuid),
('credential_cleanup',$8::uuid)) AS receipt(kind,run_id)
WHERE database.datname=pg_catalog.current_database();
INSERT INTO platform.database_generation_manifests VALUES($1,1,$4,$5);
INSERT INTO platform.release_generations VALUES
($3,$2,$1,'fresh',1,1,1,1,$4,$5,$9,'cluster_bootstrap',$6,
'extension_admin',$7,'credential_cleanup',$8);
INSERT INTO platform.extension_registry VALUES
('vector','0.8.3','platform_ext','harpy_extension_admin','fresh',$7);
RESET ROLE;`
	_, err := db.admin.Exec(ctx, sql, pgx.QueryExecModeSimpleProtocol,
		"30000000-0000-4000-8000-000000000001",
		"30000000-0000-4000-8000-000000000002",
		"30000000-0000-4000-8000-000000000003",
		migrationDigest, publicationExtensionDigest,
		"30000000-0000-4000-8000-000000000004",
		"30000000-0000-4000-8000-000000000005",
		"30000000-0000-4000-8000-000000000006",
		runtimev1.PostgresImageDigest,
	)
	if err != nil {
		t.Fatalf("seed publication startup contract: %v", err)
	}
}
