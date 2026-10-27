//go:build integration

package executor

import (
	"context"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

const pg18EnvironmentID = "77777777-7777-4777-8777-777777777777"
const pg18ReleaseID = "88888888-8888-4888-8888-888888888888"
const pg18MigrationRunID = "99999999-9999-4999-8999-999999999999"
const pg18ExtensionRunID = "12121212-1212-4212-8212-121212121212"
const pg18ExtensionDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func pg18PrepareFoundation(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	pg18ExecuteSetup(t, ctx, admin, []string{
		"SET ROLE forgecrud_owner",
		`CREATE TABLE platform.environment_identity(
singleton boolean PRIMARY KEY,environment_id uuid,database_origin text,
database_system_identifier text,database_oid bigint,release_id uuid)`,
		`CREATE TABLE platform.bootstrap_receipts(
kind text,run_id uuid,environment_id uuid,release_id uuid,database_origin text,
database_system_identifier text,database_oid bigint)`,
		`CREATE TABLE platform.extension_registry(
name text,version text,schema_name text,owner_role text,database_origin text,
verified_run_id uuid)`,
		`CREATE TABLE platform.migration_runs(
run_id uuid PRIMARY KEY,release_id uuid,environment_id uuid,database_origin text,
state text,migration_manifest_sha256 text,command text,extension_manifest_sha256 text)`,
		`CREATE TABLE platform.schema_contracts(
component text PRIMARY KEY,database_generation bigint,minimum_runtime_generation bigint,
migration_manifest_sha256 text,extension_manifest_sha256 text,status text,
verified_run_id uuid,revision bigint,verified_at timestamptz)`,
		`CREATE TABLE platform.release_generations(
migration_run_id uuid PRIMARY KEY,release_id uuid,environment_id uuid,database_origin text,
database_generation bigint,minimum_runtime_generation bigint,runtime_generation bigint,
minimum_database_generation bigint,migration_manifest_sha256 text,
extension_manifest_sha256 text,postgres_image_digest text,bootstrap_receipt_kind text,
bootstrap_receipt_run_id uuid,extension_receipt_kind text,extension_receipt_run_id uuid,
cleanup_receipt_kind text,cleanup_receipt_run_id uuid)`,
		"RESET ROLE",
	})
	pg18SeedFoundation(t, ctx, admin)
}

func pg18SeedFoundation(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	migrationDigest := pg18MigrationDigest(t)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO platform.environment_identity
SELECT true,$1,'fresh',system_identifier::text,database.oid,$2
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database
WHERE database.datname=pg_catalog.current_database()`, []any{pg18EnvironmentID, pg18ReleaseID}},
		{`INSERT INTO platform.migration_runs
VALUES($1,$2,$3,'fresh','succeeded',$4,'up',$5)`, []any{
			pg18MigrationRunID, pg18ReleaseID, pg18EnvironmentID,
			migrationDigest, pg18ExtensionDigest,
		}},
		{`INSERT INTO platform.schema_contracts
VALUES('core',1,1,$1,$2,'ready',$3,1,pg_catalog.now())`, []any{
			migrationDigest, pg18ExtensionDigest, pg18MigrationRunID,
		}},
		{pg18ReceiptSeedSQL, []any{pg18EnvironmentID, pg18ReleaseID}},
		{`INSERT INTO platform.release_generations VALUES
($1,$2,$3,'fresh',1,1,1,1,$4,$5,$6,
'cluster_bootstrap','13131313-1313-4313-8313-131313131313',
'extension_admin','12121212-1212-4212-8212-121212121212',
'credential_cleanup','14141414-1414-4414-8414-141414141414')`, []any{
			pg18MigrationRunID, pg18ReleaseID, pg18EnvironmentID,
			migrationDigest, pg18ExtensionDigest, runtimev1.PostgresImageDigest,
		}},
		{`INSERT INTO platform.extension_registry VALUES
('vector','0.8.3','platform_ext','forgecrud_extension_admin','fresh',$1)`,
			[]any{pg18ExtensionRunID}},
	}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement.query, statement.args...); err != nil {
			t.Fatalf("seed PostgreSQL foundation: %v", err)
		}
	}
}

const pg18ReceiptSeedSQL = `INSERT INTO platform.bootstrap_receipts
SELECT kind,run_id,$1,$2,'fresh',system_identifier::text,database.oid
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database,
(VALUES ('cluster_bootstrap'::text,'13131313-1313-4313-8313-131313131313'::uuid),
('extension_admin','12121212-1212-4212-8212-121212121212'),
('credential_cleanup','14141414-1414-4414-8414-141414141414')) AS receipt(kind,run_id)
WHERE database.datname=pg_catalog.current_database()`
