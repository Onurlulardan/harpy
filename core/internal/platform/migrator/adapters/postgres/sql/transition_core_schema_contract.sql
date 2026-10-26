INSERT INTO platform.schema_contracts (
    component,
    database_generation,
    minimum_runtime_generation,
    migration_manifest_sha256,
    extension_manifest_sha256,
    status,
    verified_run_id,
    revision,
    verified_at
) VALUES (
    'core',
    $1,
    $2,
    $3,
    $4,
    $6,
    $5,
    1,
    pg_catalog.statement_timestamp()
)
ON CONFLICT (component) DO UPDATE SET
    database_generation = EXCLUDED.database_generation,
    minimum_runtime_generation = EXCLUDED.minimum_runtime_generation,
    migration_manifest_sha256 = EXCLUDED.migration_manifest_sha256,
    extension_manifest_sha256 = EXCLUDED.extension_manifest_sha256,
    status = EXCLUDED.status,
    verified_run_id = EXCLUDED.verified_run_id,
    revision = platform.schema_contracts.revision + 1,
    verified_at = GREATEST(
        pg_catalog.statement_timestamp(),
        platform.schema_contracts.verified_at + interval '1 microsecond'
    )
RETURNING database_generation,
          minimum_runtime_generation,
          migration_manifest_sha256,
          extension_manifest_sha256,
          status,
          verified_run_id::text,
          revision;
