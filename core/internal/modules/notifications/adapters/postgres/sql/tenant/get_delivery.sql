SELECT id::text, message_id::text, provider_config_version_id::text,
       channel, state, attempt, maximum_attempts, lease_owner, lease_generation,
       lease_until, not_before, outcome_code, outcome_unknown, mandatory, row_version
FROM notifications_deliveries
WHERE id = $1::uuid;
