-- +goose Up
CREATE TABLE iam_system_record_imports (
    import_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    job_id uuid NOT NULL UNIQUE,
    idempotency_receipt_id uuid NOT NULL UNIQUE,
    resource_kind text NOT NULL,
    actor_id uuid NOT NULL,
    source_id uuid NOT NULL UNIQUE,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    catalog_version_id uuid NOT NULL,
    catalog_sha256 bytea NOT NULL,
    identity_manifest_sha256 bytea NOT NULL,
    request_sha256 bytea NOT NULL,
    idempotency_key_sha256 bytea NOT NULL,
    duplicate_action text NOT NULL,
    mapping_document bytea NOT NULL,
    mapping_sha256 bytea NOT NULL,
    field_count integer NOT NULL,
    state text NOT NULL DEFAULT 'waiting',
    next_row_index bigint NOT NULL DEFAULT 1,
    processed_rows bigint NOT NULL DEFAULT 0,
    created_rows bigint NOT NULL DEFAULT 0,
    updated_rows bigint NOT NULL DEFAULT 0,
    skipped_rows bigint NOT NULL DEFAULT 0,
    rejected_rows bigint NOT NULL DEFAULT 0,
    decompressed_bytes bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    started_at timestamptz,
    completed_at timestamptz,
    metadata_retain_until timestamptz NOT NULL,
    CONSTRAINT iam_system_record_imports_source_fk FOREIGN KEY (source_id)
        REFERENCES iam_system_import_sources (source_id) ON DELETE RESTRICT,
    CONSTRAINT iam_system_record_imports_identity_check CHECK (
        pg_catalog.uuid_extract_version(import_id) = 7 AND
        pg_catalog.uuid_extract_version(job_id) = 7 AND
        pg_catalog.uuid_extract_version(idempotency_receipt_id) = 7 AND
        pg_catalog.uuid_extract_version(source_id) = 7 AND
        tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        actor_id <> '00000000-0000-0000-0000-000000000000'::uuid
    ),
    CONSTRAINT iam_system_record_imports_request_check CHECK (
        resource_kind IN ('users','roles','organizations') AND
        token_version BETWEEN 1 AND 9007199254740991 AND
        authorization_version BETWEEN 1 AND 9007199254740991 AND
        authorization_clock BETWEEN 1 AND 9007199254740991 AND
        pg_catalog.uuid_extract_version(catalog_version_id) = 7 AND
        pg_catalog.octet_length(catalog_sha256) = 32 AND
        pg_catalog.octet_length(identity_manifest_sha256) = 32 AND
        pg_catalog.octet_length(request_sha256) = 32 AND
        pg_catalog.octet_length(idempotency_key_sha256) = 32 AND
        duplicate_action IN ('skip','update') AND
        pg_catalog.octet_length(mapping_document) BETWEEN 2 AND 1048576 AND
        pg_catalog.sha256(mapping_document) = mapping_sha256 AND
        pg_catalog.jsonb_typeof(
            pg_catalog.convert_from(mapping_document, 'UTF8')::jsonb
        ) = 'object' AND field_count BETWEEN 1 AND 500
    ),
    CONSTRAINT iam_system_record_imports_progress_check CHECK (
        state IN ('waiting','running','completed','cancelled','failed') AND
        next_row_index = processed_rows + 1 AND
        processed_rows BETWEEN 0 AND 1000000 AND
        created_rows BETWEEN 0 AND processed_rows AND
        updated_rows BETWEEN 0 AND processed_rows AND
        skipped_rows BETWEEN 0 AND processed_rows AND
        rejected_rows BETWEEN 0 AND 10000 AND
        created_rows + updated_rows + skipped_rows + rejected_rows = processed_rows AND
        decompressed_bytes BETWEEN 0 AND 2147483648 AND
        (state IN ('waiting','running') = (completed_at IS NULL))
    ),
    CONSTRAINT iam_system_record_imports_time_check CHECK (
        pg_catalog.isfinite(created_at) AND
        metadata_retain_until >= created_at + interval '90 days' AND
        (started_at IS NULL OR started_at >= created_at) AND
        (completed_at IS NULL OR completed_at >= created_at)
    ),
    CONSTRAINT iam_system_record_imports_idempotency_unique UNIQUE (
        actor_id, resource_kind, idempotency_key_sha256
    )
);

ALTER TABLE iam_system_import_sources
ADD CONSTRAINT iam_system_import_sources_claim_fk
FOREIGN KEY (claimed_import_id) REFERENCES iam_system_record_imports (import_id)
ON DELETE RESTRICT;

CREATE INDEX iam_system_record_imports_actor_created_idx
ON iam_system_record_imports (actor_id, created_at DESC, import_id);

CREATE INDEX iam_system_record_imports_state_idx
ON iam_system_record_imports (state, created_at, import_id);
