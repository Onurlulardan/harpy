//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func seedCatalogFoundation(t *testing.T, ctx context.Context, database *sql.DB,
	origin, extensionSchema, extensionOwner string,
) {
	t.Helper()
	digest, err := schemacontract.IntegrationMigrationManifestSHA256()
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	values := catalogFoundationValues{origin: origin, migrationSHA256: digest,
		extensionSchema: extensionSchema, extensionOwner: extensionOwner}
	for _, command := range values.commands() {
		if _, err = transaction.ExecContext(ctx, command.statement, command.arguments...); err != nil {
			t.Fatal(err)
		}
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

type catalogFoundationValues struct {
	origin, migrationSHA256, extensionSchema, extensionOwner string
}

type catalogFoundationCommand struct {
	statement string
	arguments []any
}

func (value catalogFoundationValues) commands() []catalogFoundationCommand {
	const environmentID = "78000000-0000-4700-8700-000000000001"
	const releaseID = "78000000-0000-4700-8700-000000000002"
	const migrationRunID = "78000000-0000-4700-8700-000000000003"
	const bootstrapID = "78100000-0000-4710-8710-000000000001"
	const extensionID = "78100000-0000-4710-8710-000000000002"
	const cleanupID = "78100000-0000-4710-8710-000000000003"
	return []catalogFoundationCommand{
		{catalogEnvironmentIdentitySQL, []any{environmentID, value.origin, releaseID}},
		{catalogMigrationRunSQL, []any{migrationRunID, releaseID, environmentID,
			value.origin, value.migrationSHA256, catalogExtensionDigest}},
		{catalogSchemaContractSQL, []any{value.migrationSHA256,
			catalogExtensionDigest, migrationRunID}},
		{catalogBootstrapReceiptsSQL, []any{environmentID, releaseID, value.origin,
			bootstrapID, extensionID, cleanupID}},
		{catalogReleaseGenerationSQL, []any{migrationRunID, releaseID, environmentID,
			value.origin, value.migrationSHA256, catalogExtensionDigest,
			runtimev1.PostgresImageDigest, bootstrapID, extensionID, cleanupID}},
		{catalogExtensionRegistrySQL, []any{value.extensionSchema, value.extensionOwner,
			value.origin, extensionID}},
	}
}

const catalogEnvironmentIdentitySQL = `INSERT INTO platform.environment_identity
SELECT true,$1,$2,system_identifier::text,database.oid,$3
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database
WHERE database.datname=pg_catalog.current_database()`

const catalogMigrationRunSQL = `INSERT INTO platform.migration_runs
VALUES($1,$2,$3,$4,'succeeded',$5,'up',$6)`

const catalogSchemaContractSQL = `INSERT INTO platform.schema_contracts
VALUES('core',1,1,$1,$2,'ready',$3,1,pg_catalog.now())`

const catalogBootstrapReceiptsSQL = `INSERT INTO platform.bootstrap_receipts
SELECT kind,run_id,$1,$2,$3,system_identifier::text,database.oid
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database,
(VALUES ('cluster_bootstrap'::text,$4::uuid),('extension_admin',$5::uuid),
 ('credential_cleanup',$6::uuid)) AS receipt(kind,run_id)
WHERE database.datname=pg_catalog.current_database()`

const catalogReleaseGenerationSQL = `INSERT INTO platform.release_generations VALUES(
$1,$2,$3,$4,1,1,1,1,$5,$6,$7,
'cluster_bootstrap',$8,'extension_admin',$9,'credential_cleanup',$10)`

const catalogExtensionRegistrySQL = `INSERT INTO platform.extension_registry
VALUES('vector','0.8.3',$1,$2,$3,$4)`
