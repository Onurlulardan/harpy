SELECT recipient_id::text, category, channel, enabled, explicit, verified_target,
       quiet_timezone, quiet_start_minute, quiet_end_minute, row_version
FROM notifications_preferences
WHERE recipient_id = $1::uuid
ORDER BY category, channel;
