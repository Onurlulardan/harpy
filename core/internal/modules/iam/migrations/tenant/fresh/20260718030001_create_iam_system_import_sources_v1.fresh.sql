-- +goose Up
CREATE TABLE iam_system_import_sources (
    source_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    idempotency_receipt_id uuid NOT NULL UNIQUE,
    resource_kind text NOT NULL,
    actor_id uuid NOT NULL,
    idempotency_key_sha256 bytea NOT NULL,
    format text NOT NULL,
    object_key text NOT NULL,
    object_version text NOT NULL,
    header_sha256 bytea NOT NULL,
    ciphertext_sha256 bytea NOT NULL,
    ciphertext_bytes bigint NOT NULL,
    plaintext_sha256 bytea NOT NULL,
    plaintext_bytes bigint NOT NULL,
    key_id text NOT NULL,
    wrapped_dek bytea NOT NULL,
    wrap_nonce bytea NOT NULL,
    provider_request_id text NOT NULL,
    uploaded_at timestamptz NOT NULL,
    state text NOT NULL DEFAULT 'ready',
    claimed_import_id uuid,
    claimed_at timestamptz,
    delete_after timestamptz NOT NULL,
    deleted_at timestamptz,
    CONSTRAINT iam_system_import_sources_identity_check CHECK (
        pg_catalog.uuid_extract_version(source_id) = 7 AND
        pg_catalog.uuid_extract_version(idempotency_receipt_id) = 7 AND
        tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT iam_system_import_sources_object_check CHECK (
        resource_kind IN ('users','roles','organizations') AND
        pg_catalog.octet_length(idempotency_key_sha256) = 32 AND
        format IN ('csv','jsonl') AND
        pg_catalog.octet_length(object_key) BETWEEN 1 AND 512 AND
        pg_catalog.octet_length(object_version) BETWEEN 1 AND 512 AND
        pg_catalog.octet_length(header_sha256) = 32 AND
        pg_catalog.octet_length(ciphertext_sha256) = 32 AND
        ciphertext_bytes BETWEEN 1 AND 538968064 AND
        pg_catalog.octet_length(plaintext_sha256) = 32 AND
        plaintext_bytes BETWEEN 1 AND 536870912 AND
        pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
        pg_catalog.octet_length(wrapped_dek) BETWEEN 32 AND 1024 AND
        pg_catalog.octet_length(wrap_nonce) = 12 AND
        pg_catalog.octet_length(provider_request_id) BETWEEN 1 AND 256
    ),
    CONSTRAINT iam_system_import_sources_state_check CHECK (
        (state = 'ready' AND claimed_import_id IS NULL AND claimed_at IS NULL AND
         deleted_at IS NULL) OR
        (state = 'claimed' AND pg_catalog.uuid_extract_version(claimed_import_id) = 7 AND
         claimed_at IS NOT NULL AND deleted_at IS NULL) OR
        (state = 'deleted' AND deleted_at IS NOT NULL)
    ),
    CONSTRAINT iam_system_import_sources_time_check CHECK (
        pg_catalog.isfinite(uploaded_at) AND delete_after = uploaded_at + interval '24 hours' AND
        (claimed_at IS NULL OR claimed_at >= uploaded_at) AND
        (deleted_at IS NULL OR deleted_at >= uploaded_at)
    ),
    CONSTRAINT iam_system_import_sources_idempotency_unique UNIQUE (
        actor_id, resource_kind, idempotency_key_sha256
    ),
    CONSTRAINT iam_system_import_sources_object_unique UNIQUE (
        object_key, object_version
    )
);

CREATE INDEX iam_system_import_sources_cleanup_idx
ON iam_system_import_sources (delete_after, source_id)
WHERE state IN ('ready','deleted');
