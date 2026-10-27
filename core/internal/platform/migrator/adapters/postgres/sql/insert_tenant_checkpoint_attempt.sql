INSERT INTO platform.tenant_migration_checkpoint_attempts (
    tenant_id, owner, logical_version, attempt, database_origin,
    selected_path, selected_sha256, run_id, state,
    started_at, finished_at, error_code, recorded_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp(), $10,
    pg_catalog.clock_timestamp()
);
