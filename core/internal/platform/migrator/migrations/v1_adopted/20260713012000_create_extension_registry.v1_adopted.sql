-- +goose Up
CREATE TABLE platform.extension_registry (
    name text NOT NULL CHECK (name ~ '^[a-z][a-z0-9_]{0,62}$'),
    version text NOT NULL CHECK (version ~ '^[0-9]+\.[0-9]+(\.[0-9]+)?$'),
    schema_name text NOT NULL CHECK (schema_name ~ '^[a-z][a-z0-9_]{0,62}$'),
    owner_role text NOT NULL CHECK (owner_role ~ '^[a-z][a-z0-9_]{0,62}$'),
    control_file_sha256 text NOT NULL CHECK (control_file_sha256 ~ '^[0-9a-f]{64}$'),
    capabilities text[] NOT NULL CHECK (cardinality(capabilities) BETWEEN 1 AND 64),
    database_origin text NOT NULL CHECK (database_origin IN ('fresh', 'v1_adopted')),
    verified_run_id uuid NOT NULL,
    registered_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    PRIMARY KEY (name, version, verified_run_id),
    UNIQUE (name, verified_run_id)
);
