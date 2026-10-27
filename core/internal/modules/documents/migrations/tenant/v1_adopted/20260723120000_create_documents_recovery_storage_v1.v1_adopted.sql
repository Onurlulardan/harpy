-- +goose Up
CREATE TABLE documents_dr_holds (
    recovery_set_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_set_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(tenant_set_sha256) = 32),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'HOLDING','COPYING','VERIFIED','RELEASED','FAILED'
    )),
    inventory_sha256 bytea CHECK (
        inventory_sha256 IS NULL OR pg_catalog.octet_length(inventory_sha256) = 32
    ),
    inventory_count bigint CHECK (inventory_count IS NULL OR inventory_count >= 0),
    failure_code text CHECK (
        failure_code IS NULL OR failure_code ~ '^[a-z0-9_]{1,64}$'
    ),
    started_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    released_at timestamptz,
    CHECK ((inventory_sha256 IS NULL) = (inventory_count IS NULL)),
    CHECK ((lifecycle_state = 'FAILED') = (failure_code IS NOT NULL)),
    CHECK ((lifecycle_state = 'RELEASED') = (released_at IS NOT NULL))
);

CREATE TABLE documents_dr_copy_receipts (
    recovery_set_id uuid NOT NULL
        REFERENCES documents_dr_holds(recovery_set_id) ON DELETE RESTRICT,
    version_id uuid NOT NULL REFERENCES documents_versions(id) ON DELETE RESTRICT,
    source_object_key text NOT NULL
        CHECK (pg_catalog.octet_length(source_object_key) BETWEEN 1 AND 1024),
    source_object_version text NOT NULL
        CHECK (pg_catalog.octet_length(source_object_version) BETWEEN 1 AND 1024),
    target_object_key text NOT NULL
        CHECK (pg_catalog.octet_length(target_object_key) BETWEEN 1 AND 1024),
    target_object_version text NOT NULL
        CHECK (pg_catalog.octet_length(target_object_version) BETWEEN 1 AND 1024),
    ciphertext_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(ciphertext_sha256) = 32),
    ciphertext_bytes bigint NOT NULL CHECK (ciphertext_bytes > 0),
    key_id text NOT NULL CHECK (key_id ~ '^[a-z0-9][a-z0-9_.-]{0,63}$'),
    verified_at timestamptz NOT NULL,
    PRIMARY KEY (recovery_set_id, version_id),
    UNIQUE (recovery_set_id, target_object_key, target_object_version)
);

CREATE TABLE documents_restore_locator_remaps (
    restore_operation_id uuid NOT NULL,
    recovery_set_id uuid NOT NULL
        REFERENCES documents_dr_holds(recovery_set_id) ON DELETE RESTRICT,
    version_id uuid NOT NULL REFERENCES documents_versions(id) ON DELETE RESTRICT,
    source_object_key text NOT NULL,
    source_object_version text NOT NULL,
    target_object_key text NOT NULL,
    target_object_version text NOT NULL,
    target_object_generation bigint NOT NULL CHECK (target_object_generation > 0),
    ciphertext_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(ciphertext_sha256) = 32),
    applied_at timestamptz NOT NULL,
    PRIMARY KEY (restore_operation_id, version_id),
    UNIQUE (restore_operation_id, target_object_key, target_object_version)
);

CREATE TABLE documents_restore_transient_dispositions (
    restore_operation_id uuid NOT NULL,
    target_kind text NOT NULL CHECK (target_kind IN ('UPLOAD','ARCHIVE')),
    target_id uuid NOT NULL,
    disposition text NOT NULL CHECK (disposition IN ('EXPIRED','CANCELLED')),
    disposed_at timestamptz NOT NULL,
    PRIMARY KEY (restore_operation_id, target_kind, target_id)
);
