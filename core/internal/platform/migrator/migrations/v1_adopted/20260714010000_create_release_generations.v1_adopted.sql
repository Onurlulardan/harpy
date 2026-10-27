-- +goose Up
CREATE TABLE platform.release_generations (
    migration_run_id uuid PRIMARY KEY
        REFERENCES platform.migration_runs(run_id) ON DELETE RESTRICT,
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    database_origin text NOT NULL
        CHECK (database_origin IN ('fresh', 'v1_adopted')),
    database_generation bigint NOT NULL CHECK (database_generation > 0),
    minimum_runtime_generation bigint NOT NULL
        CHECK (minimum_runtime_generation > 0),
    runtime_generation bigint NOT NULL CHECK (runtime_generation > 0),
    minimum_database_generation bigint NOT NULL
        CHECK (minimum_database_generation > 0),
    migration_manifest_sha256 text NOT NULL
        CHECK (migration_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    extension_manifest_sha256 text NOT NULL
        CHECK (extension_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    postgres_image_digest text NOT NULL
        CHECK (postgres_image_digest ~ '^sha256:[0-9a-f]{64}$'),
    bootstrap_receipt_kind text NOT NULL DEFAULT 'cluster_bootstrap'
        CHECK (bootstrap_receipt_kind = 'cluster_bootstrap'),
    bootstrap_receipt_run_id uuid NOT NULL,
    extension_receipt_kind text NOT NULL DEFAULT 'extension_admin'
        CHECK (extension_receipt_kind = 'extension_admin'),
    extension_receipt_run_id uuid NOT NULL,
    cleanup_receipt_kind text NOT NULL DEFAULT 'credential_cleanup'
        CHECK (cleanup_receipt_kind = 'credential_cleanup'),
    cleanup_receipt_run_id uuid NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    CONSTRAINT release_generations_environment_release_unique
        UNIQUE (environment_id, release_id),
    CONSTRAINT release_generations_runtime_tuple_unique UNIQUE (
        environment_id,
        runtime_generation,
        minimum_database_generation,
        migration_manifest_sha256,
        extension_manifest_sha256,
        postgres_image_digest
    ),
    CONSTRAINT release_generations_bootstrap_receipt_fk FOREIGN KEY (
        bootstrap_receipt_kind,
        bootstrap_receipt_run_id
    ) REFERENCES platform.bootstrap_receipts(kind, run_id) ON DELETE RESTRICT,
    CONSTRAINT release_generations_extension_receipt_fk FOREIGN KEY (
        extension_receipt_kind,
        extension_receipt_run_id
    ) REFERENCES platform.bootstrap_receipts(kind, run_id) ON DELETE RESTRICT,
    CONSTRAINT release_generations_cleanup_receipt_fk FOREIGN KEY (
        cleanup_receipt_kind,
        cleanup_receipt_run_id
    ) REFERENCES platform.bootstrap_receipts(kind, run_id) ON DELETE RESTRICT,
    CONSTRAINT release_generations_database_manifest_fk FOREIGN KEY (
        environment_id,
        database_generation,
        migration_manifest_sha256,
        extension_manifest_sha256
    ) REFERENCES platform.database_generation_manifests (
        environment_id,
        database_generation,
        migration_manifest_sha256,
        extension_manifest_sha256
    ) ON DELETE RESTRICT,
    CONSTRAINT release_generations_receipts_distinct CHECK (
        bootstrap_receipt_run_id <> extension_receipt_run_id
        AND bootstrap_receipt_run_id <> cleanup_receipt_run_id
        AND extension_receipt_run_id <> cleanup_receipt_run_id
    ),
    CONSTRAINT release_generations_database_floor CHECK (
        minimum_database_generation <= database_generation
    ),
    CONSTRAINT release_generations_runtime_floor CHECK (
        minimum_runtime_generation <= runtime_generation
    )
);
