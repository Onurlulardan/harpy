-- +goose Up
CREATE TABLE platform.bootstrap_receipts (
    kind text NOT NULL
        CHECK (kind IN ('cluster_bootstrap', 'extension_admin', 'credential_cleanup')),
    run_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    release_id uuid NOT NULL,
    database_origin text NOT NULL CHECK (database_origin IN ('fresh', 'v1_adopted')),
    database_system_identifier text NOT NULL
        CHECK (database_system_identifier ~ '^[1-9][0-9]{0,19}$'),
    database_oid bigint NOT NULL CHECK (database_oid BETWEEN 1 AND 4294967295),
    operation_digest text NOT NULL CHECK (operation_digest ~ '^[0-9a-f]{64}$'),
    canonical_document bytea NOT NULL
        CHECK (octet_length(canonical_document) BETWEEN 2 AND 4194304),
    canonical_sha256 text NOT NULL UNIQUE CHECK (canonical_sha256 ~ '^[0-9a-f]{64}$'),
    signature_key_id text NOT NULL CHECK (signature_key_id ~ '^[a-z][a-z0-9_-]{2,63}$'),
    issued_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    PRIMARY KEY (kind, run_id),
    CHECK (issued_at <= recorded_at)
);
