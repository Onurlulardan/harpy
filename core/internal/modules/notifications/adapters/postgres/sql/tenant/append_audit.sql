INSERT INTO notifications_audit_entries (
    id, action, resource_id, actor_id, resource_revision, occurred_at
) VALUES (
    pg_catalog.uuidv7(), $1, $2::uuid, $3::uuid, $4,
    pg_catalog.transaction_timestamp()
);
