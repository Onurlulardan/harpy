-- +goose Up
CREATE TABLE platform.migration_runs (
    run_id uuid PRIMARY KEY,
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL CHECK (database_origin IN ('fresh', 'v1_adopted')),
    command text NOT NULL CHECK (command IN ('plan', 'verify', 'up', 'schema_executor')),
    state text NOT NULL CHECK (state IN ('planned', 'running', 'succeeded', 'failed', 'fenced')),
    migration_manifest_sha256 text NOT NULL CHECK (migration_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    extension_manifest_sha256 text NOT NULL CHECK (extension_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    roles_manifest_sha256 text NOT NULL CHECK (roles_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    privileges_manifest_sha256 text NOT NULL CHECK (privileges_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    ownership_manifest_sha256 text NOT NULL CHECK (ownership_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    started_at timestamptz,
    finished_at timestamptz,
    failure_code text CHECK (failure_code ~ '^[a-z][a-z0-9_]{2,63}$'),
    CHECK ((state = 'planned' AND started_at IS NULL AND finished_at IS NULL AND failure_code IS NULL)
        OR (state = 'running' AND started_at IS NOT NULL AND finished_at IS NULL AND failure_code IS NULL)
        OR (state = 'succeeded' AND started_at IS NOT NULL AND finished_at IS NOT NULL AND failure_code IS NULL)
        OR (state IN ('failed', 'fenced') AND finished_at IS NOT NULL AND failure_code IS NOT NULL)),
    CHECK (started_at IS NULL OR created_at <= started_at),
    CHECK (finished_at IS NULL OR started_at <= finished_at)
);
