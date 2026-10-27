INSERT INTO notifications_adoption_checkpoints (
    run_id, source_table, last_source_id, adopted_count, source_checksum,
    target_checksum, updated_at
) VALUES (
    $1::uuid, 'notifications', $2::uuid, $3, $4::bytea, $5::bytea,
    pg_catalog.transaction_timestamp()
)
ON CONFLICT (run_id, source_table) DO UPDATE
SET last_source_id = EXCLUDED.last_source_id,
    adopted_count = EXCLUDED.adopted_count,
    source_checksum = EXCLUDED.source_checksum,
    target_checksum = EXCLUDED.target_checksum,
    updated_at = EXCLUDED.updated_at;
