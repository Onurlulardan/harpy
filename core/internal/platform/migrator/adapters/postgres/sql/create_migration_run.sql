INSERT INTO platform.migration_runs (
    run_id,
    release_id,
    environment_id,
    database_origin,
    command,
    state,
    migration_manifest_sha256,
    extension_manifest_sha256,
    roles_manifest_sha256,
    privileges_manifest_sha256,
    ownership_manifest_sha256
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    'planned',
    $6,
    $7,
    $8,
    $9,
    $10
)
ON CONFLICT (run_id) DO NOTHING;
