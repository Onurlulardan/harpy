UPDATE notifications_deliveries
SET state = 'sending',
    attempt = attempt + 1,
    lease_owner = $2,
    lease_generation = lease_generation + 1,
    lease_until = $3::timestamptz + interval '1 minute',
    row_version = row_version + 1
WHERE id = $1::uuid AND state = 'pending' AND outcome_unknown = false
  AND not_before <= $3 AND attempt < maximum_attempts
RETURNING id::text, message_id::text, provider_config_version_id::text,
          channel, state, attempt, maximum_attempts, lease_owner, lease_generation,
          lease_until, not_before, outcome_code, outcome_unknown, mandatory, row_version;
