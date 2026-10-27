//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const pipelineEnvironmentID = "91000000-0000-4000-8000-000000000001"
const pipelineReleaseID = "91000000-0000-4000-8000-000000000002"
const pipelineMigrationRunID = "91000000-0000-4000-8000-000000000003"
const pipelineExtensionRunID = "91000000-0000-4000-8000-000000000004"

func preparePipelineFoundation(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	if _, err := database.ExecContext(ctx, pipelineFoundationTables); err != nil {
		t.Fatalf("create pipeline runtime foundation: %v", err)
	}
	digest := pipelineMigrationDigest(t)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO platform.environment_identity
SELECT true,$1,'fresh',system_identifier::text,database.oid::bigint,$2
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database
WHERE database.datname=pg_catalog.current_database()`, []any{pipelineEnvironmentID, pipelineReleaseID}},
		{`INSERT INTO platform.migration_runs VALUES
($1,$2,$3,'fresh','succeeded','up',$4,$5)`, []any{
			pipelineMigrationRunID, pipelineReleaseID, pipelineEnvironmentID,
			digest, pipelineExtensionDigest,
		}},
		{`INSERT INTO platform.schema_contracts VALUES
('core',1,1,$1,$2,'ready',$3,1,pg_catalog.clock_timestamp())`, []any{
			digest, pipelineExtensionDigest, pipelineMigrationRunID,
		}},
		{pipelineBootstrapReceiptsSQL, []any{pipelineEnvironmentID, pipelineReleaseID}},
		{`INSERT INTO platform.database_generation_manifests VALUES($1,1,$2,$3)`, []any{
			pipelineEnvironmentID, digest, pipelineExtensionDigest,
		}},
		{pipelineReleaseGenerationSQL, []any{
			pipelineMigrationRunID, pipelineReleaseID, pipelineEnvironmentID,
			digest, pipelineExtensionDigest, runtimev1.PostgresImageDigest,
		}},
		{`INSERT INTO platform.extension_registry VALUES
('vector','0.8.3','platform_ext','harpy_extension_admin','fresh',$1)`, []any{
			pipelineExtensionRunID,
		}},
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatalf("seed pipeline runtime foundation: %v", err)
		}
	}
	if _, err := database.ExecContext(
		ctx, "GRANT SELECT ON ALL TABLES IN SCHEMA platform TO harpy_runtime",
	); err != nil {
		t.Fatalf("grant pipeline foundation reads: %v", err)
	}
}

const pipelineBootstrapReceiptsSQL = `INSERT INTO platform.bootstrap_receipts
SELECT kind,run_id,$1,$2,'fresh',system_identifier::text,database.oid::bigint
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database,
(VALUES
 ('cluster_bootstrap'::text,'91000000-0000-4000-8000-000000000005'::uuid),
 ('extension_admin','91000000-0000-4000-8000-000000000004'),
 ('credential_cleanup','91000000-0000-4000-8000-000000000006')
) AS receipt(kind,run_id)
WHERE database.datname=pg_catalog.current_database()`

const pipelineReleaseGenerationSQL = `INSERT INTO platform.release_generations VALUES
($1,$2,$3,'fresh',1,1,1,1,$4,$5,$6,
 'cluster_bootstrap','91000000-0000-4000-8000-000000000005',
 'extension_admin','91000000-0000-4000-8000-000000000004',
 'credential_cleanup','91000000-0000-4000-8000-000000000006')`
