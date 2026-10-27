//go:build integration

package postgres

const pipelineFoundationTables = `
CREATE TABLE platform.environment_identity (
    singleton boolean PRIMARY KEY,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL,
    database_system_identifier text NOT NULL,
    database_oid bigint NOT NULL,
    release_id uuid NOT NULL
);
CREATE TABLE platform.bootstrap_receipts (
    kind text NOT NULL,
    run_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    release_id uuid NOT NULL,
    database_origin text NOT NULL,
    database_system_identifier text NOT NULL,
    database_oid bigint NOT NULL
);
CREATE TABLE platform.extension_registry (
    name text NOT NULL,
    version text NOT NULL,
    schema_name text NOT NULL,
    owner_role text NOT NULL,
    database_origin text NOT NULL,
    verified_run_id uuid NOT NULL
);
CREATE TABLE platform.migration_runs (
    run_id uuid PRIMARY KEY,
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL,
    state text NOT NULL,
    command text NOT NULL,
    migration_manifest_sha256 text NOT NULL,
    extension_manifest_sha256 text NOT NULL
);
CREATE TABLE platform.schema_contracts (
    component text PRIMARY KEY,
    database_generation bigint NOT NULL,
    minimum_runtime_generation bigint NOT NULL,
    migration_manifest_sha256 text NOT NULL,
    extension_manifest_sha256 text NOT NULL,
    status text NOT NULL,
    verified_run_id uuid NOT NULL,
    revision bigint NOT NULL,
    verified_at timestamptz NOT NULL
);
CREATE TABLE platform.database_generation_manifests (
    environment_id uuid NOT NULL,
    database_generation bigint NOT NULL,
    migration_manifest_sha256 text NOT NULL,
    extension_manifest_sha256 text NOT NULL,
    PRIMARY KEY (environment_id,database_generation),
    UNIQUE (
        environment_id,database_generation,
        migration_manifest_sha256,extension_manifest_sha256
    )
);
CREATE TABLE platform.release_generations (
    migration_run_id uuid PRIMARY KEY,
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL,
    database_generation bigint NOT NULL,
    minimum_runtime_generation bigint NOT NULL,
    runtime_generation bigint NOT NULL,
    minimum_database_generation bigint NOT NULL,
    migration_manifest_sha256 text NOT NULL,
    extension_manifest_sha256 text NOT NULL,
    postgres_image_digest text NOT NULL,
    bootstrap_receipt_kind text NOT NULL,
    bootstrap_receipt_run_id uuid NOT NULL,
    extension_receipt_kind text NOT NULL,
    extension_receipt_run_id uuid NOT NULL,
    cleanup_receipt_kind text NOT NULL,
    cleanup_receipt_run_id uuid NOT NULL,
    FOREIGN KEY (
        environment_id,database_generation,
        migration_manifest_sha256,extension_manifest_sha256
    ) REFERENCES platform.database_generation_manifests (
        environment_id,database_generation,
        migration_manifest_sha256,extension_manifest_sha256
    )
);`
