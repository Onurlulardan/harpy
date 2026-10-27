WITH inserted AS (
    INSERT INTO documents_adoption_runs (
        run_id, cutover_generation, manifest_sha256, transform_version,
        lifecycle_state, source_count, target_count, source_plaintext_bytes,
        target_plaintext_bytes, pending_object_count, pending_job_count,
        created_at, updated_at
    )
    VALUES ($1::uuid, $2, $3, $4, 'INVENTORY', 0, 0, 0, 0, 0, 0, $5, $5)
    ON CONFLICT (run_id) DO NOTHING
    RETURNING *
), result AS (
    SELECT r.*, false AS replay FROM inserted r
    UNION ALL
    SELECT r.*, true AS replay FROM documents_adoption_runs r
    WHERE r.run_id = $1::uuid AND NOT EXISTS (SELECT 1 FROM inserted)
      AND r.cutover_generation = $2 AND r.manifest_sha256 = $3
      AND r.transform_version = $4
)
SELECT run_id, cutover_generation, manifest_sha256, lifecycle_state,
       source_count, target_count, source_plaintext_bytes, target_plaintext_bytes,
       source_merkle_sha256, target_merkle_sha256, pending_object_count,
       pending_job_count, final_lsn::text, applied_lsn::text, acceptance_sha256,
       created_at, updated_at, replay
FROM result;
