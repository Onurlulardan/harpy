INSERT INTO platform.database_generation_manifests (
    environment_id,
    database_generation,
    migration_manifest_sha256,
    extension_manifest_sha256
) VALUES ($1, $2, $3, $4)
ON CONFLICT (environment_id, database_generation) DO NOTHING;
