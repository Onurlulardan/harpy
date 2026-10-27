WITH run AS MATERIALIZED (
    SELECT run_id FROM documents_adoption_runs
    WHERE run_id = $1::uuid AND cutover_generation = $2
      AND manifest_sha256 = $3 AND lifecycle_state IN ('INVENTORY','COPYING','SHADOW')
    FOR UPDATE
), changed AS (
    UPDATE documents_blob_copy_checkpoints c
    SET lifecycle_state = 'COPYING', updated_at = $5
    FROM run WHERE c.run_id = run.run_id AND c.source_version_id = $4::uuid
      AND c.lifecycle_state IN ('INVENTORIED','COPYING')
    RETURNING c.*
), run_changed AS (
    UPDATE documents_adoption_runs r SET lifecycle_state = 'COPYING', updated_at = $5
    FROM changed c WHERE r.run_id = c.run_id AND r.lifecycle_state = 'INVENTORY'
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
