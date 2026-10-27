INSERT INTO platform.platform_migration_history (
    version,
    record_kind,
    is_applied,
    owner,
    scope,
    origin,
    selected_path,
    selected_sha256,
    run_id,
    target_database_generation,
    next_minimum_runtime_generation,
    applied_at
) VALUES (
    $1,
    'migration',
    TRUE,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    pg_catalog.statement_timestamp()
);
