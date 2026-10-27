-- +goose Up
CREATE TABLE documents_archives (
    id uuid PRIMARY KEY,
    actor_id uuid NOT NULL,
    authorization_revision bigint NOT NULL CHECK (authorization_revision > 0),
    job_id uuid NOT NULL UNIQUE,
    retry_of_archive_id uuid,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    billing_reservation_id uuid NOT NULL UNIQUE,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'PENDING','READY','EXPIRED','PURGED'
    )),
    entry_count integer NOT NULL CHECK (entry_count BETWEEN 0 AND 10000),
    entries_sha256 bytea CHECK (
        entries_sha256 IS NULL OR pg_catalog.octet_length(entries_sha256) = 32
    ),
    plaintext_bytes bigint NOT NULL CHECK (plaintext_bytes BETWEEN 0 AND 10737418240),
    ciphertext_bytes bigint CHECK (ciphertext_bytes IS NULL OR ciphertext_bytes > 0),
    artifact_sha256 bytea CHECK (
        artifact_sha256 IS NULL OR pg_catalog.octet_length(artifact_sha256) = 32
    ),
    object_key text CHECK (
        object_key IS NULL OR pg_catalog.octet_length(object_key) BETWEEN 1 AND 1024
    ),
    object_version text CHECK (
        object_version IS NULL OR pg_catalog.octet_length(object_version) BETWEEN 1 AND 1024
    ),
    object_generation bigint CHECK (object_generation IS NULL OR object_generation > 0),
    key_id text CHECK (key_id IS NULL OR key_id ~ '^[a-z0-9][a-z0-9_.-]{0,63}$'),
    wrapped_dek bytea CHECK (
        wrapped_dek IS NULL OR pg_catalog.octet_length(wrapped_dek) BETWEEN 115 AND 256
    ),
    ready_at timestamptz,
    expires_at timestamptz,
    delete_receipt_sha256 bytea CHECK (
        delete_receipt_sha256 IS NULL OR pg_catalog.octet_length(delete_receipt_sha256) = 32
    ),
    purged_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT documents_archives_retry_fk FOREIGN KEY (retry_of_archive_id)
        REFERENCES documents_archives(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    UNIQUE (actor_id, request_sha256),
    CHECK ((lifecycle_state = 'PENDING' AND ready_at IS NULL) OR
           (lifecycle_state <> 'PENDING' AND entry_count > 0 AND entries_sha256 IS NOT NULL AND
            ciphertext_bytes IS NOT NULL AND artifact_sha256 IS NOT NULL AND
            object_key IS NOT NULL AND object_version IS NOT NULL AND
            object_generation IS NOT NULL AND key_id IS NOT NULL AND wrapped_dek IS NOT NULL AND
            ready_at IS NOT NULL AND expires_at = ready_at + interval '24 hours')),
    CHECK ((lifecycle_state = 'PURGED' AND delete_receipt_sha256 IS NOT NULL AND
            purged_at IS NOT NULL) OR lifecycle_state <> 'PURGED')
);

CREATE TABLE documents_archive_entries (
    archive_id uuid NOT NULL,
    entry_order integer NOT NULL CHECK (entry_order BETWEEN 0 AND 9999),
    document_id uuid NOT NULL,
    version_id uuid NOT NULL,
    plaintext_bytes bigint NOT NULL CHECK (plaintext_bytes BETWEEN 1 AND 104857600),
    plaintext_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(plaintext_sha256) = 32),
    added_at timestamptz NOT NULL,
    PRIMARY KEY (archive_id, entry_order),
    CONSTRAINT documents_archive_entries_archive_fk FOREIGN KEY (archive_id)
        REFERENCES documents_archives(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT documents_archive_entries_version_fk FOREIGN KEY (version_id)
        REFERENCES documents_versions(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    UNIQUE (archive_id, version_id)
);

CREATE INDEX documents_archives_job_idx ON documents_archives(job_id);
CREATE INDEX documents_archives_expiry_idx
    ON documents_archives(expires_at, id) WHERE lifecycle_state = 'READY';

