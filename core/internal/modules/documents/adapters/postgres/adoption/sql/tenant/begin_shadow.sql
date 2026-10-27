WITH changed AS (
    UPDATE documents_adoption_runs r
    SET lifecycle_state = 'SHADOW', updated_at = $4
    WHERE r.run_id = $1::uuid AND r.cutover_generation = $2
      AND r.manifest_sha256 = $3 AND r.lifecycle_state = 'COPYING'
      AND NOT EXISTS (
          SELECT 1 FROM documents_blob_copy_checkpoints c
          WHERE c.run_id = r.run_id
            AND c.lifecycle_state NOT IN ('VERIFIED','SOURCE_HOLD_RELEASED')
      )
    RETURNING r.*
)
SELECT run_id, cutover_generation, manifest_sha256, lifecycle_state,
       source_count, target_count, source_plaintext_bytes, target_plaintext_bytes,
       source_merkle_sha256, target_merkle_sha256, pending_object_count,
       pending_job_count, final_lsn::text, applied_lsn::text, acceptance_sha256,
       created_at, updated_at, false
FROM changed;
