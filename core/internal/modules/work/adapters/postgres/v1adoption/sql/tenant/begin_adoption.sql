INSERT INTO work_adoption_runs (
    run_id, cutover_generation, manifest_sha256, transform_version, state,
    source_count, target_count, source_sha256, target_sha256,
    final_lsn, applied_lsn, acceptance_sha256, created_at, updated_at
) VALUES (
    $1, $2, $3, 'work_adoption_v1', 'INVENTORY', $4, $5, $6, $7,
    pg_current_wal_lsn(), NULL, NULL, $8, $8
)
ON CONFLICT (run_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
WHERE work_adoption_runs.cutover_generation = EXCLUDED.cutover_generation
  AND work_adoption_runs.manifest_sha256 = EXCLUDED.manifest_sha256
  AND work_adoption_runs.source_count = EXCLUDED.source_count
  AND work_adoption_runs.target_count = EXCLUDED.target_count
  AND work_adoption_runs.source_sha256 = EXCLUDED.source_sha256
  AND work_adoption_runs.target_sha256 = EXCLUDED.target_sha256
  AND work_adoption_runs.state IN ('INVENTORY','COPYING','RECONCILING','BLOCKED')
