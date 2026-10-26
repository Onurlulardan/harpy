WITH candidate AS (
    SELECT delivery.id,
           (
               delivery.mandatory
               OR (
                   COALESCE(
                       preference.enabled,
                       delivery.channel = 'email'
                           AND message.category IN ('security','workflow_action','billing')
                   )
                   AND (
                       delivery.channel <> 'telegram'
                       OR (
                           COALESCE(preference.explicit, false)
                           AND COALESCE(preference.verified_target, false)
                       )
                   )
               )
           ) AS allowed
    FROM notifications_deliveries delivery
    JOIN notifications_messages message ON message.id = delivery.message_id
    LEFT JOIN notifications_preferences preference
      ON preference.recipient_id = message.recipient_id
     AND preference.category = message.category
     AND preference.channel = delivery.channel
    WHERE delivery.id = $1::uuid
      AND delivery.state = 'pending'
      AND delivery.outcome_unknown = false
      AND delivery.not_before <= $3
      AND delivery.attempt < delivery.maximum_attempts
    FOR UPDATE OF delivery
), decided AS (
    UPDATE notifications_deliveries delivery
    SET state = CASE WHEN candidate.allowed THEN 'sending' ELSE 'suppressed' END,
        attempt = delivery.attempt + CASE WHEN candidate.allowed THEN 1 ELSE 0 END,
        lease_owner = CASE WHEN candidate.allowed THEN $2 ELSE '' END,
        lease_generation = delivery.lease_generation
            + CASE WHEN candidate.allowed THEN 1 ELSE 0 END,
        lease_until = CASE WHEN candidate.allowed
            THEN $3::timestamptz + interval '1 minute' ELSE NULL END,
        outcome_code = CASE WHEN candidate.allowed THEN '' ELSE 'PREFERENCE_DISABLED' END,
        row_version = delivery.row_version + 1
    FROM candidate
    WHERE delivery.id = candidate.id
    RETURNING delivery.*
)
SELECT id::text, message_id::text, provider_config_version_id::text,
       channel, state, attempt, maximum_attempts, lease_owner, lease_generation,
       lease_until, not_before, outcome_code, outcome_unknown, mandatory, row_version
FROM decided;
