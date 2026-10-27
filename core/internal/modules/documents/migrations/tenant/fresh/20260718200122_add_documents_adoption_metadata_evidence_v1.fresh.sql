-- +goose Up
ALTER TABLE documents_blob_copy_checkpoints
    ADD COLUMN source_metadata_sha256 bytea
        CHECK (source_metadata_sha256 IS NULL OR
               pg_catalog.octet_length(source_metadata_sha256) = 32),
    ADD COLUMN source_lsn pg_lsn,
    ADD COLUMN target_object_generation bigint
        CHECK (target_object_generation IS NULL OR target_object_generation > 0),
    ADD COLUMN target_ciphertext_sha256 bytea
        CHECK (target_ciphertext_sha256 IS NULL OR
               pg_catalog.octet_length(target_ciphertext_sha256) = 32),
    ADD COLUMN target_key_id text
        CHECK (target_key_id IS NULL OR target_key_id ~ '^[a-z0-9][a-z0-9_.-]{0,63}$'),
    ADD COLUMN target_wrapped_dek bytea
        CHECK (target_wrapped_dek IS NULL OR
               pg_catalog.octet_length(target_wrapped_dek) BETWEEN 115 AND 256);

ALTER TABLE documents_blob_copy_checkpoints
    ADD CONSTRAINT documents_blob_copy_metadata_evidence_v1
    CHECK (lifecycle_state NOT IN ('VERIFIED','SOURCE_HOLD_RELEASED') OR
           (source_metadata_sha256 IS NOT NULL AND
            metadata_sha256 = source_metadata_sha256 AND
            source_lsn IS NOT NULL AND applied_lsn = source_lsn AND
            target_object_generation IS NOT NULL AND
            target_ciphertext_sha256 IS NOT NULL AND target_key_id IS NOT NULL AND
           target_wrapped_dek IS NOT NULL));

GRANT INSERT (source_metadata_sha256, source_lsn, target_object_generation, target_ciphertext_sha256, target_key_id, target_wrapped_dek)
ON documents_blob_copy_checkpoints TO harpy_adopter;
GRANT UPDATE (source_metadata_sha256, source_lsn, target_object_generation, target_ciphertext_sha256, target_key_id, target_wrapped_dek)
ON documents_blob_copy_checkpoints TO harpy_adopter;
