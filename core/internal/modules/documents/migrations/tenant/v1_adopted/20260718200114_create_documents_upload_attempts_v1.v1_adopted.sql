-- +goose Up
ALTER TABLE documents_items
    ADD COLUMN reference_epoch bigint NOT NULL DEFAULT 1 CHECK (reference_epoch > 0);

CREATE TABLE documents_upload_attempts (
    upload_id uuid NOT NULL,
    attempt_generation bigint NOT NULL CHECK (attempt_generation > 0),
    effect_receipt_id uuid NOT NULL UNIQUE,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN (
        'RECEIVING','FAILED','EFFECT_RECORDED','FINALIZED','CLEANUP_PENDING','CLEANED'
    )),
    object_key text,
    object_version text,
    object_generation bigint,
    plaintext_bytes bigint,
    ciphertext_bytes bigint,
    plaintext_sha256 bytea,
    ciphertext_sha256 bytea,
    detected_mime text,
    file_kind text,
    key_id text,
    wrapped_dek bytea,
    scan_signature_revision text,
    scanned_at timestamptz,
    cleanup_deletion_id uuid,
    cleanup_receipt_sha256 bytea,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (upload_id, attempt_generation),
    CONSTRAINT documents_upload_attempts_upload_fk FOREIGN KEY (upload_id)
        REFERENCES documents_uploads(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT documents_upload_attempts_cleanup_fk FOREIGN KEY (cleanup_deletion_id)
        REFERENCES documents_blob_deletions(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK ((lifecycle_state IN ('RECEIVING','FAILED') AND
            object_key IS NULL AND object_version IS NULL AND
            object_generation IS NULL AND plaintext_bytes IS NULL AND ciphertext_bytes IS NULL AND
            plaintext_sha256 IS NULL AND ciphertext_sha256 IS NULL AND detected_mime IS NULL AND
            file_kind IS NULL AND key_id IS NULL AND wrapped_dek IS NULL AND
            scan_signature_revision IS NULL AND scanned_at IS NULL) OR
        (lifecycle_state IN ('EFFECT_RECORDED','FINALIZED') AND
         object_key IS NOT NULL AND object_version IS NOT NULL AND object_generation > 0 AND
         plaintext_bytes > 0 AND ciphertext_bytes > plaintext_bytes AND
         pg_catalog.octet_length(plaintext_sha256) = 32 AND
         pg_catalog.octet_length(ciphertext_sha256) = 32 AND
         key_id IS NOT NULL AND wrapped_dek IS NOT NULL AND
         detected_mime IS NOT NULL AND file_kind IS NOT NULL AND
         scan_signature_revision IS NOT NULL AND scanned_at IS NOT NULL) OR
        (lifecycle_state IN ('CLEANUP_PENDING','CLEANED') AND
         object_key IS NOT NULL AND object_version IS NOT NULL AND object_generation > 0 AND
         plaintext_bytes > 0 AND ciphertext_bytes > plaintext_bytes)),
    CHECK ((lifecycle_state IN ('CLEANUP_PENDING','CLEANED')) =
        (cleanup_deletion_id IS NOT NULL)),
    CHECK ((lifecycle_state = 'CLEANED') =
        (cleanup_receipt_sha256 IS NOT NULL AND
         pg_catalog.octet_length(cleanup_receipt_sha256) = 32))
);

CREATE INDEX documents_upload_attempts_cleanup_idx
    ON documents_upload_attempts(lifecycle_state, updated_at, upload_id, attempt_generation)
    WHERE lifecycle_state IN ('CLEANUP_PENDING','CLEANED');

GRANT UPDATE (reference_epoch) ON documents_items TO harpy_runtime;
GRANT INSERT (reference_epoch) ON documents_items TO harpy_adopter;
GRANT UPDATE (reference_epoch) ON documents_items TO harpy_adopter;

GRANT SELECT ON documents_upload_attempts TO harpy_runtime;
GRANT INSERT (upload_id, attempt_generation, effect_receipt_id, lifecycle_state, created_at, updated_at)
ON documents_upload_attempts TO harpy_runtime;
GRANT UPDATE (lifecycle_state, object_key, object_version, object_generation, plaintext_bytes, ciphertext_bytes, plaintext_sha256, ciphertext_sha256, detected_mime, file_kind, key_id, wrapped_dek, scan_signature_revision, scanned_at, cleanup_deletion_id, cleanup_receipt_sha256, updated_at)
ON documents_upload_attempts TO harpy_runtime;
GRANT SELECT ON documents_upload_attempts TO harpy_adopter;
GRANT INSERT (upload_id, attempt_generation, effect_receipt_id, lifecycle_state, object_key, object_version, object_generation, plaintext_bytes, ciphertext_bytes, plaintext_sha256, ciphertext_sha256, detected_mime, file_kind, key_id, wrapped_dek, scan_signature_revision, scanned_at, cleanup_deletion_id, cleanup_receipt_sha256, created_at, updated_at)
ON documents_upload_attempts TO harpy_adopter;
GRANT UPDATE (lifecycle_state, object_key, object_version, object_generation, plaintext_bytes, ciphertext_bytes, plaintext_sha256, ciphertext_sha256, detected_mime, file_kind, key_id, wrapped_dek, scan_signature_revision, scanned_at, cleanup_deletion_id, cleanup_receipt_sha256, updated_at)
ON documents_upload_attempts TO harpy_adopter;
