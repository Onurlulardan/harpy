-- +goose Up
CREATE TABLE documents_operations (
    operation_id uuid PRIMARY KEY,
    operation_kind text NOT NULL
        CHECK (pg_catalog.octet_length(operation_kind) BETWEEN 1 AND 128),
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL
        CHECK (pg_catalog.octet_length(idempotency_key) BETWEEN 1 AND 256),
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    aggregate_kind text NOT NULL
        CHECK (aggregate_kind IN ('FOLDER','ITEM','UPLOAD','ARCHIVE','LINK')),
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
    response_status smallint NOT NULL CHECK (response_status BETWEEN 200 AND 299),
    response_document bytea NOT NULL
        CHECK (pg_catalog.octet_length(response_document) <= 65536),
    occurred_at timestamptz NOT NULL,
    UNIQUE (actor_id, operation_kind, idempotency_key)
);

CREATE TABLE documents_kek_releases (
    key_id text PRIMARY KEY CHECK (key_id ~ '^[a-z0-9][a-z0-9_.-]{0,63}$'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('CURRENT','RETAINED','RETIRED')),
    key_material_sha256 bytea NOT NULL
        CHECK (pg_catalog.octet_length(key_material_sha256) = 32),
    reference_count bigint NOT NULL CHECK (reference_count >= 0),
    activated_at timestamptz NOT NULL,
    retire_after timestamptz,
    retired_at timestamptz,
    retirement_evidence_sha256 bytea CHECK (
        retirement_evidence_sha256 IS NULL OR
        pg_catalog.octet_length(retirement_evidence_sha256) = 32
    ),
    CHECK ((lifecycle_state = 'CURRENT' AND retire_after IS NULL AND retired_at IS NULL) OR
           (lifecycle_state = 'RETAINED' AND retire_after IS NOT NULL AND retired_at IS NULL) OR
           (lifecycle_state = 'RETIRED' AND reference_count = 0 AND retired_at IS NOT NULL AND
            retirement_evidence_sha256 IS NOT NULL))
);

CREATE UNIQUE INDEX documents_kek_current_key
    ON documents_kek_releases((lifecycle_state)) WHERE lifecycle_state = 'CURRENT';

CREATE TABLE documents_format_releases (
    format_id text PRIMARY KEY CHECK (format_id ~ '^FCDOC[0-9]+$'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('CURRENT','RETAINED','RETIRED')),
    decoder_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(decoder_sha256) = 32),
    reference_count bigint NOT NULL CHECK (reference_count >= 0),
    last_reference_purged_at timestamptz,
    retire_after timestamptz,
    retired_at timestamptz,
    retirement_evidence_sha256 bytea CHECK (
        retirement_evidence_sha256 IS NULL OR
        pg_catalog.octet_length(retirement_evidence_sha256) = 32
    ),
    CHECK ((lifecycle_state <> 'RETIRED') OR
           (reference_count = 0 AND last_reference_purged_at IS NOT NULL AND
            retire_after = last_reference_purged_at + interval '24 hours' AND
            retired_at >= retire_after AND retirement_evidence_sha256 IS NOT NULL))
);

CREATE TABLE documents_audit_entries (
    audit_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    action text NOT NULL CHECK (pg_catalog.octet_length(action) BETWEEN 1 AND 128),
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
    reason text CHECK (reason IS NULL OR pg_catalog.octet_length(reason) <= 4096),
    occurred_at timestamptz NOT NULL,
    CONSTRAINT documents_audit_entries_operation_fk FOREIGN KEY (operation_id)
        REFERENCES documents_operations(operation_id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

