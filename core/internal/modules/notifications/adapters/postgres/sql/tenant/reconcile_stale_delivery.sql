WITH stale AS (
    UPDATE notifications_deliveries
    SET state = 'failed', lease_owner = '', lease_until = NULL,
        outcome_code = 'NOTIFICATION_OUTCOME_UNKNOWN', outcome_unknown = true,
        row_version = row_version + 1
    WHERE id = $1::uuid AND state = 'sending' AND lease_until <= $2::timestamptz
    RETURNING id, channel, attempt, lease_generation, lease_until
), recorded AS (
    INSERT INTO notifications_delivery_attempts (
        id, delivery_id, attempt_number, lease_generation, provider_kind,
        response_class, provider_reference_digest, status_code, error_code,
        started_at, finished_at
    )
    SELECT pg_catalog.uuidv7(), id, attempt, lease_generation,
           CASE WHEN channel = 'email' THEN 'smtp' ELSE 'telegram' END,
           'stale_lease_outcome_unknown', decode(repeat('00', 32), 'hex'), '',
           'NOTIFICATION_OUTCOME_UNKNOWN', $2::timestamptz - interval '1 minute', $2
    FROM stale
    RETURNING delivery_id
)
SELECT stale.id::text
FROM stale
JOIN recorded ON recorded.delivery_id = stale.id;
