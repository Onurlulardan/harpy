INSERT INTO notifications_adoption_issues (
    id, run_id, source_table, source_key_digest, code, created_at, resolved_at
) VALUES (
    $1::uuid, $2::uuid, 'notifications', $3::bytea, $4,
    pg_catalog.transaction_timestamp(), NULL
)
ON CONFLICT (id) DO NOTHING;
