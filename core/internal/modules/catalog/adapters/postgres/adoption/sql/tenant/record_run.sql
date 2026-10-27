WITH inserted AS (
    INSERT INTO catalog_adoption_runs (
        run_id, source_manifest_sha256, transform_sha256, source_count, target_count,
        source_root_sha256, target_root_sha256, state, blocker_code,
        cutover_generation, created_at, verified_at
    ) VALUES (
        $1::uuid, $2::text, $3::text, $4::bigint, $5::bigint,
        $6::text, $7::text, $8::text, $9::text,
        $10::bigint, $11::timestamptz, $12::timestamptz
    ) ON CONFLICT DO NOTHING
    RETURNING run_id, source_manifest_sha256, transform_sha256, source_count,
              target_count, source_root_sha256, target_root_sha256, state,
              blocker_code, cutover_generation, created_at, verified_at
)
SELECT * FROM inserted
UNION ALL
SELECT run_id, source_manifest_sha256, transform_sha256, source_count,
       target_count, source_root_sha256, target_root_sha256, state,
       blocker_code, cutover_generation, created_at, verified_at
FROM catalog_adoption_runs
WHERE run_id = $1::uuid
LIMIT 1
