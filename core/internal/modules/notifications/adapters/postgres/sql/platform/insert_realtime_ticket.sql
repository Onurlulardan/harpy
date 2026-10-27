INSERT INTO platform.notification_realtime_tickets (
    id, tenant_id, user_id, token_hash, authorization_version,
    origin, created_at, expires_at, consumed_at
) VALUES (
    $1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, NULL
);
