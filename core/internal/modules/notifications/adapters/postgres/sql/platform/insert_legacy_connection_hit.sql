INSERT INTO platform.notification_legacy_connection_hits (
    id, tenant_id, user_id, consumer_manifest_digest, occurred_at
) VALUES (
    pg_catalog.uuidv7(), $1::uuid, $2::uuid, $3::bytea,
    pg_catalog.transaction_timestamp()
);
