INSERT INTO notifications_preferences (
    recipient_id, category, channel, enabled, explicit, verified_target,
    quiet_timezone, quiet_start_minute, quiet_end_minute, row_version, updated_at
)
SELECT $1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, pg_catalog.transaction_timestamp()
WHERE $11::bigint = 0
ON CONFLICT (recipient_id, category, channel) DO UPDATE
SET enabled = EXCLUDED.enabled,
    explicit = EXCLUDED.explicit,
    verified_target = EXCLUDED.verified_target,
    quiet_timezone = EXCLUDED.quiet_timezone,
    quiet_start_minute = EXCLUDED.quiet_start_minute,
    quiet_end_minute = EXCLUDED.quiet_end_minute,
    row_version = EXCLUDED.row_version,
    updated_at = EXCLUDED.updated_at
WHERE notifications_preferences.row_version = $11
RETURNING row_version;
