WITH run AS MATERIALIZED (
    SELECT run_id FROM documents_adoption_runs
    WHERE run_id = $1::uuid AND cutover_generation = $2
      AND manifest_sha256 = $3 AND lifecycle_state IN ('INVENTORY','COPYING','SHADOW')
    FOR UPDATE
), inserted AS (
    INSERT INTO documents_blob_copy_checkpoints (
        run_id, source_document_id, source_version_id, source_object_key,
        source_object_revision, source_bytes, source_sha256,
        source_metadata_sha256, source_lsn, target_document_id, target_version_id,
        lifecycle_state, updated_at
    )
    SELECT run_id, $4::uuid, $5::uuid, $6, $7, $8, $9, $10, $11::pg_lsn,
           $12::uuid, $13::uuid, 'INVENTORIED', $14 FROM run
    ON CONFLICT (run_id, source_version_id, source_object_revision) DO NOTHING
    RETURNING *
), result AS (
    SELECT c.*, false AS replay FROM inserted c
    UNION ALL
    SELECT c.*, true AS replay FROM documents_blob_copy_checkpoints c
    JOIN run ON run.run_id = c.run_id
    WHERE c.source_version_id = $5::uuid AND c.source_object_revision = $7
      AND c.source_document_id = $4::uuid AND c.source_object_key = $6
      AND c.source_bytes = $8 AND c.source_sha256 = $9
      AND c.source_metadata_sha256 = $10 AND c.source_lsn = $11::pg_lsn
      AND c.target_document_id = $12::uuid AND c.target_version_id = $13::uuid
)
SELECT run_id, source_document_id, source_version_id, source_object_key,
       source_object_revision, source_bytes, source_sha256, source_metadata_sha256,
       source_lsn::text,
       target_document_id, target_version_id, target_object_key,
       target_object_version, target_object_generation, target_ciphertext_bytes,
       target_plaintext_sha256, target_ciphertext_sha256, target_key_id,
       target_wrapped_dek, metadata_sha256, applied_lsn::text, lifecycle_state,
       issue_code, verified_at, updated_at, replay
FROM result;
