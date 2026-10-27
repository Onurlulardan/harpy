-- +goose Up
CREATE INDEX release_generations_compatibility_lookup
ON platform.release_generations (
    environment_id,
    database_origin,
    migration_manifest_sha256,
    extension_manifest_sha256,
    runtime_generation,
    minimum_database_generation
);
