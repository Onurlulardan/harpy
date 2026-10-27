-- +goose Up
CREATE TABLE iam_system_record_exports (
    export_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    job_id uuid NOT NULL UNIQUE,
    idempotency_receipt_id uuid NOT NULL UNIQUE,
    resource_kind text NOT NULL,
    actor_id uuid NOT NULL,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    catalog_version_id uuid NOT NULL,
    catalog_sha256 bytea NOT NULL,
    query_document bytea NOT NULL,
    query_sha256 bytea NOT NULL,
    request_sha256 bytea NOT NULL,
    idempotency_key_sha256 bytea NOT NULL,
    format text NOT NULL,
    state text NOT NULL DEFAULT 'waiting',
    processed_rows bigint NOT NULL DEFAULT 0,
    plaintext_bytes bigint NOT NULL DEFAULT 0,
    object_key text,
    object_version text,
    header_sha256 bytea,
    ciphertext_sha256 bytea,
    ciphertext_bytes bigint,
    key_id text,
    wrapped_dek bytea,
    wrap_nonce bytea,
    attempt_generation bigint,
    created_at timestamptz NOT NULL,
    started_at timestamptz,
    ready_at timestamptz,
    expires_at timestamptz,
    completed_at timestamptz,
    metadata_retain_until timestamptz NOT NULL,
    CONSTRAINT iam_system_record_exports_identity_check CHECK (
        pg_catalog.uuid_extract_version(export_id) = 7 AND
        pg_catalog.uuid_extract_version(job_id) = 7 AND
        pg_catalog.uuid_extract_version(idempotency_receipt_id) = 7 AND
        tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT iam_system_record_exports_request_check CHECK (
        resource_kind IN ('users','roles','organizations') AND
        token_version BETWEEN 1 AND 9007199254740991 AND
        authorization_version BETWEEN 1 AND 9007199254740991 AND
        authorization_clock BETWEEN 1 AND 9007199254740991 AND
        pg_catalog.uuid_extract_version(catalog_version_id) = 7 AND
        pg_catalog.octet_length(catalog_sha256) = 32 AND
        pg_catalog.octet_length(query_document) BETWEEN 2 AND 262144 AND
        pg_catalog.sha256(query_document) = query_sha256 AND
        pg_catalog.jsonb_typeof(
            pg_catalog.convert_from(query_document, 'UTF8')::jsonb
        ) = 'object' AND
        pg_catalog.octet_length(request_sha256) = 32 AND
        pg_catalog.octet_length(idempotency_key_sha256) = 32 AND
        format IN ('csv','jsonl','xlsx')
    ),
    CONSTRAINT iam_system_record_exports_progress_check CHECK (
        processed_rows BETWEEN 0 AND 1000000 AND
        plaintext_bytes BETWEEN 0 AND 2147483648
    ),
    CONSTRAINT iam_system_record_exports_artifact_check CHECK (
        (state IN ('waiting','running','cancelled','failed') AND
         object_key IS NULL AND object_version IS NULL AND header_sha256 IS NULL AND
         ciphertext_sha256 IS NULL AND ciphertext_bytes IS NULL AND key_id IS NULL AND
         wrapped_dek IS NULL AND wrap_nonce IS NULL AND attempt_generation IS NULL AND
         ready_at IS NULL AND expires_at IS NULL) OR
        (state IN ('ready','expired') AND
         pg_catalog.octet_length(object_key) BETWEEN 1 AND 512 AND
         pg_catalog.octet_length(object_version) BETWEEN 1 AND 512 AND
         pg_catalog.octet_length(header_sha256) = 32 AND
         pg_catalog.octet_length(ciphertext_sha256) = 32 AND
         ciphertext_bytes BETWEEN 1 AND 2148532224 AND
         pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
         pg_catalog.octet_length(wrapped_dek) BETWEEN 32 AND 1024 AND
         pg_catalog.octet_length(wrap_nonce) = 12 AND
         attempt_generation BETWEEN 1 AND 9007199254740991 AND
         ready_at IS NOT NULL AND expires_at = ready_at + interval '24 hours')
    ),
    CONSTRAINT iam_system_record_exports_time_check CHECK (
        pg_catalog.isfinite(created_at) AND
        metadata_retain_until >= created_at + interval '90 days' AND
        (started_at IS NULL OR started_at >= created_at) AND
        (ready_at IS NULL OR ready_at >= started_at) AND
        (completed_at IS NULL OR completed_at >= created_at)
    ),
    CONSTRAINT iam_system_record_exports_idempotency_unique UNIQUE (
        actor_id, resource_kind, idempotency_key_sha256
    )
);

CREATE INDEX iam_system_record_exports_actor_created_idx
ON iam_system_record_exports (actor_id, created_at DESC, export_id);

CREATE INDEX iam_system_record_exports_state_idx
ON iam_system_record_exports (state, created_at, export_id);
