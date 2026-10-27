-- +goose Up
CREATE TABLE platform.schema_contracts (
    component text PRIMARY KEY CHECK (component ~ '^[a-z][a-z0-9_-]{0,63}$'),
    database_generation bigint NOT NULL CHECK (database_generation > 0),
    minimum_runtime_generation bigint NOT NULL CHECK (minimum_runtime_generation > 0),
    migration_manifest_sha256 text NOT NULL CHECK (migration_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    extension_manifest_sha256 text NOT NULL CHECK (extension_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    status text NOT NULL CHECK (status IN ('ready', 'fenced')),
    verified_run_id uuid NOT NULL REFERENCES platform.migration_runs(run_id) ON DELETE RESTRICT,
    revision bigint NOT NULL CHECK (revision > 0),
    verified_at timestamptz NOT NULL
);
