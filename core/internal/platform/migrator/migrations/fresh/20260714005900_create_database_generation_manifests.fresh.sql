-- +goose Up
CREATE TABLE platform.database_generation_manifests (
    environment_id uuid NOT NULL
        REFERENCES platform.environment_identity(environment_id) ON DELETE RESTRICT,
    database_generation bigint NOT NULL CHECK (database_generation > 0),
    migration_manifest_sha256 text NOT NULL
        CHECK (migration_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    extension_manifest_sha256 text NOT NULL
        CHECK (extension_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    PRIMARY KEY (environment_id, database_generation),
    CONSTRAINT database_generation_manifests_exact_key UNIQUE (
        environment_id,
        database_generation,
        migration_manifest_sha256,
        extension_manifest_sha256
    )
);
