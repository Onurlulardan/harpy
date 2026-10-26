//go:build integration

package postgres

import (
	"context"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/jackc/pgx/v5"
)

const runtimeEnvironmentID = "22222222-2222-4222-8222-222222222222"
const runtimeReleaseID = "33333333-3333-4333-8333-333333333333"
const runtimeMigrationRunID = "11111111-1111-4111-8111-111111111111"
const runtimeExtensionRunID = "44444444-4444-4444-8444-444444444444"

func seedRuntimeFoundation(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	migrationDigest := runtimeMigrationDigest(t)
	statements := []string{
		`INSERT INTO platform.environment_identity
SELECT true,$1,'fresh',system_identifier::text,database.oid::bigint,$2
FROM pg_catalog.pg_control_system(), pg_catalog.pg_database AS database
WHERE database.datname=pg_catalog.current_database()`,
		`INSERT INTO platform.migration_runs VALUES
($1,$2,$3,'fresh','succeeded','up',$4,$5)`,
		`INSERT INTO platform.schema_contracts VALUES
		('core',1,1,$1,$2,'ready',$3,1,pg_catalog.now())`,
		`INSERT INTO platform.bootstrap_receipts
SELECT kind,run_id,$1,$2,'fresh',system_identifier::text,database.oid::bigint
FROM pg_catalog.pg_control_system(), pg_catalog.pg_database AS database,
(VALUES ('cluster_bootstrap'::text,'55555555-5555-4555-8555-555555555555'::uuid),
('extension_admin','44444444-4444-4444-8444-444444444444'),
('credential_cleanup','66666666-6666-4666-8666-666666666666')) AS receipt(kind,run_id)
WHERE database.datname=pg_catalog.current_database()`,
		`INSERT INTO platform.database_generation_manifests VALUES($1,1,$2,$3)`,
		`INSERT INTO platform.release_generations VALUES
($1,$2,$3,'fresh',1,1,1,1,$4,$5,$6,
'cluster_bootstrap','55555555-5555-4555-8555-555555555555',
'extension_admin','44444444-4444-4444-8444-444444444444',
'credential_cleanup','66666666-6666-4666-8666-666666666666')`,
		`INSERT INTO platform.extension_registry VALUES
('vector','0.8.3','platform_ext','harpy_extension_admin','fresh',$1)`,
	}
	arguments := [][]any{
		{runtimeEnvironmentID, runtimeReleaseID},
		{runtimeMigrationRunID, runtimeReleaseID, runtimeEnvironmentID,
			migrationDigest, runtimeExtensionDigest},
		{migrationDigest, runtimeExtensionDigest, runtimeMigrationRunID},
		{runtimeEnvironmentID, runtimeReleaseID},
		{runtimeEnvironmentID, migrationDigest, runtimeExtensionDigest},
		{runtimeMigrationRunID, runtimeReleaseID, runtimeEnvironmentID,
			migrationDigest, runtimeExtensionDigest, runtimev1.PostgresImageDigest},
		{runtimeExtensionRunID},
	}
	for index, statement := range statements {
		if _, err := admin.Exec(ctx, statement, arguments[index]...); err != nil {
			t.Fatalf("seed runtime foundation: %v", err)
		}
	}
}
