INSERT INTO notifications_deliveries (
    id, message_id, provider_config_version_id, channel, state, attempt,
    maximum_attempts, lease_owner, lease_generation, lease_until, not_before,
    outcome_code, outcome_unknown, mandatory, row_version
) VALUES (
    $1::uuid, $2::uuid, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15
);
