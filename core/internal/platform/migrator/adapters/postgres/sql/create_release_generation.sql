INSERT INTO platform.release_generations (
    migration_run_id,
    release_id,
    environment_id,
    database_origin,
    database_generation,
    minimum_runtime_generation,
    runtime_generation,
    minimum_database_generation,
    migration_manifest_sha256,
    extension_manifest_sha256,
    postgres_image_digest,
    bootstrap_receipt_run_id,
    extension_receipt_run_id,
    cleanup_receipt_run_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14
)
ON CONFLICT (migration_run_id) DO NOTHING;
