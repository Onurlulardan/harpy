INSERT INTO notifications_adoption_runs (
    id, source_generation, source_manifest_digest, lifecycle_state,
    started_at, completed_at, row_version
)
SELECT $1::uuid, $2, $3::bytea, 'planned',
       pg_catalog.transaction_timestamp(), NULL, 1
WHERE EXISTS (
    SELECT 1 FROM notifications_writer_authority
    WHERE singleton AND writer = 'v1' AND lifecycle_state = 'writable'
)
ON CONFLICT (source_generation, source_manifest_digest) DO NOTHING;
