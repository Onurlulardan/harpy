INSERT INTO notifications_deliveries (
    id, message_id, provider_config_version_id, channel, state, attempt,
    maximum_attempts, lease_owner, lease_generation, lease_until,
    not_before, outcome_code, outcome_unknown, mandatory, row_version
) VALUES (
    $1::uuid, $2::uuid, NULL, 'email', 'accepted', 1, 8, '', 1, NULL,
    $3, 'V1_ACCEPTED', false, false, 2
);
