WITH run AS MATERIALIZED (
    SELECT run_id FROM documents_adoption_runs
    WHERE run_id = $1::uuid AND cutover_generation = $2
      AND manifest_sha256 = $3 AND lifecycle_state IN ('COPYING','SHADOW')
    FOR UPDATE
), changed AS (
    UPDATE documents_blob_copy_checkpoints c
    SET target_object_key = $7, target_object_version = $8,
        target_object_generation = $9, target_ciphertext_bytes = $10,
        target_plaintext_sha256 = $11, target_ciphertext_sha256 = $12,
        target_key_id = $13, target_wrapped_dek = $14,
        metadata_sha256 = $15, applied_lsn = $16::pg_lsn,
        lifecycle_state = 'VERIFIED', verified_at = $17, updated_at = $17
    FROM run WHERE c.run_id = run.run_id AND c.source_version_id = $4::uuid
      AND c.source_object_revision = $5 AND c.target_version_id = $6::uuid
      AND c.lifecycle_state = 'COPYING' AND c.source_sha256 = $11
      AND c.source_metadata_sha256 = $15 AND c.source_lsn = $16::pg_lsn
    RETURNING c.*
)
SELECT run_id, source_document_id, source_version_id, source_object_key,
       source_object_revision, source_bytes, source_sha256, source_metadata_sha256,
       source_lsn::text,
       target_document_id, target_version_id, target_object_key,
       target_object_version, target_object_generation, target_ciphertext_bytes,
       target_plaintext_sha256, target_ciphertext_sha256, target_key_id,
       target_wrapped_dek, metadata_sha256, applied_lsn::text, lifecycle_state,
       issue_code, verified_at, updated_at, false
FROM changed;
