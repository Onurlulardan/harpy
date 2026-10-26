//go:build integration

package postgres

import (
	"context"
	"testing"
)

func prepareSchemaExecutorTrigramRegistry(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	origin string,
) {
	t.Helper()
	statements := []struct {
		sql  string
		args []any
	}{
		{`CREATE SCHEMA platform_ext AUTHORIZATION harpy_owner`, nil},
		{`CREATE EXTENSION pg_trgm WITH SCHEMA platform_ext`, nil},
		{`INSERT INTO platform.migration_runs(
 run_id,release_id,environment_id,database_origin,command,state,
 migration_manifest_sha256,extension_manifest_sha256,roles_manifest_sha256,
 privileges_manifest_sha256,ownership_manifest_sha256)
VALUES ($1,$2,$3,$4,'plan','planned',repeat('0',64),repeat('1',64),
        repeat('2',64),repeat('3',64),repeat('4',64))`, []any{
			schemaExecutorIntegrationUUID(898), schemaExecutorIntegrationUUID(899),
			schemaExecutorIntegrationUUID(900), origin,
		}},
		{`INSERT INTO platform.extension_registry(
 name,version,schema_name,owner_role,control_file_sha256,capabilities,
 database_origin,verified_run_id)
SELECT 'pg_trgm',extension.extversion,'platform_ext','harpy_owner',
       repeat('5',64),ARRAY['trigram_search']::text[],$2,$1
FROM pg_catalog.pg_extension AS extension WHERE extension.extname='pg_trgm'`, []any{
			schemaExecutorIntegrationUUID(898), origin,
		}},
		{`INSERT INTO platform.schema_contracts(
 component,database_generation,minimum_runtime_generation,
 migration_manifest_sha256,extension_manifest_sha256,status,
 verified_run_id,revision,verified_at)
VALUES ('core',1,1,repeat('0',64),repeat('1',64),'ready',$1,1,
        pg_catalog.clock_timestamp())`, []any{schemaExecutorIntegrationUUID(898)}},
	}
	for _, statement := range statements {
		if _, err := environment.pool.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatalf("prepare schema executor trigram registry: %v", err)
		}
	}
}
