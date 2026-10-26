WITH changed AS (
    UPDATE documents_adoption_runs r
    SET lifecycle_state = 'ACCEPTED', acceptance_sha256 = $4, updated_at = $6
    WHERE r.run_id = $1::uuid AND r.cutover_generation = $2
      AND r.manifest_sha256 = $3 AND r.lifecycle_state = 'QUIESCED'
      AND r.source_count = r.target_count
      AND r.source_plaintext_bytes = r.target_plaintext_bytes
      AND r.source_merkle_sha256 = r.target_merkle_sha256
      AND r.pending_object_count = 0 AND r.pending_job_count = 0
      AND r.final_lsn = $5::pg_lsn AND r.applied_lsn = $5::pg_lsn
    RETURNING r.*
)
SELECT run_id, cutover_generation, manifest_sha256, lifecycle_state,
       source_count, target_count, source_plaintext_bytes, target_plaintext_bytes,
       source_merkle_sha256, target_merkle_sha256, pending_object_count,
       pending_job_count, final_lsn::text, applied_lsn::text, acceptance_sha256,
       created_at, updated_at, false
FROM changed;
