WITH locked AS MATERIALIZED (
    SELECT r.* FROM documents_adoption_runs r
    WHERE r.run_id = $1::uuid AND r.cutover_generation = $2
      AND r.manifest_sha256 = $3 AND r.lifecycle_state IN ('SHADOW','RECONCILING')
    FOR UPDATE
), stats AS MATERIALIZED (
    SELECT count(*)::bigint AS source_count,
           count(*) FILTER (WHERE c.lifecycle_state IN
               ('VERIFIED','SOURCE_HOLD_RELEASED'))::bigint AS target_count,
           coalesce(sum(c.source_bytes), 0)::bigint AS source_bytes,
           coalesce(sum(c.source_bytes) FILTER (WHERE c.lifecycle_state IN
               ('VERIFIED','SOURCE_HOLD_RELEASED')), 0)::bigint AS target_bytes,
           pg_catalog.sha256(pg_catalog.convert_to(coalesce(pg_catalog.string_agg(
               c.source_version_id::text || ':' || c.source_object_revision || ':' ||
               pg_catalog.encode(c.source_sha256, 'hex') || ':' ||
               pg_catalog.encode(c.source_metadata_sha256, 'hex'), '|'
               ORDER BY c.source_version_id, c.source_object_revision), ''), 'UTF8')) AS source_root,
           pg_catalog.sha256(pg_catalog.convert_to(coalesce(pg_catalog.string_agg(
               c.source_version_id::text || ':' || c.source_object_revision || ':' ||
               pg_catalog.encode(c.target_plaintext_sha256, 'hex') || ':' ||
               pg_catalog.encode(c.metadata_sha256, 'hex'), '|'
               ORDER BY c.source_version_id, c.source_object_revision)
               FILTER (WHERE c.lifecycle_state IN
                   ('VERIFIED','SOURCE_HOLD_RELEASED')), ''), 'UTF8')) AS target_root,
           count(*) FILTER (WHERE c.lifecycle_state NOT IN
               ('VERIFIED','SOURCE_HOLD_RELEASED'))::bigint AS pending_objects,
           bool_and(c.applied_lsn IS NOT NULL AND c.applied_lsn <= $5::pg_lsn)
               FILTER (WHERE c.lifecycle_state IN
                   ('VERIFIED','SOURCE_HOLD_RELEASED')) AS lsn_valid
    FROM locked r LEFT JOIN documents_blob_copy_checkpoints c ON c.run_id = r.run_id
), changed AS (
    UPDATE documents_adoption_runs r
    SET lifecycle_state = CASE WHEN s.source_count = s.target_count
            AND s.source_bytes = s.target_bytes AND s.source_root = s.target_root
            AND s.pending_objects = 0 AND $4::pg_lsn = $5::pg_lsn
            AND $6::bigint = 0 AND coalesce(s.lsn_valid, true)
        THEN 'QUIESCED' ELSE 'RECONCILING' END,
        source_count = s.source_count, target_count = s.target_count,
        source_plaintext_bytes = s.source_bytes, target_plaintext_bytes = s.target_bytes,
        source_merkle_sha256 = s.source_root, target_merkle_sha256 = s.target_root,
        pending_object_count = s.pending_objects, pending_job_count = $6,
        final_lsn = $4::pg_lsn, applied_lsn = $5::pg_lsn, updated_at = $7
    FROM locked, stats s WHERE r.run_id = locked.run_id
    RETURNING r.*
)
SELECT run_id, cutover_generation, manifest_sha256, lifecycle_state,
       source_count, target_count, source_plaintext_bytes, target_plaintext_bytes,
       source_merkle_sha256, target_merkle_sha256, pending_object_count,
       pending_job_count, final_lsn::text, applied_lsn::text, acceptance_sha256,
       created_at, updated_at, false
FROM changed;
