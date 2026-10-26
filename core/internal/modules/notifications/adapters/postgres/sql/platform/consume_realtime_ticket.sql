UPDATE platform.notification_realtime_tickets
SET consumed_at = $6
WHERE token_hash = $1
  AND tenant_id = $2::uuid
  AND user_id = $3::uuid
  AND authorization_version = $4
  AND origin = $5
  AND consumed_at IS NULL
  AND created_at <= $6
  AND expires_at > $6
RETURNING id::text, tenant_id::text, user_id::text, token_hash,
          authorization_version, origin, created_at, expires_at, consumed_at;
