WITH current_delivery AS (
    SELECT id, channel, state
    FROM notifications_deliveries
    WHERE id = $2::uuid
    FOR UPDATE
), inserted_receipt AS (
    INSERT INTO notifications_delivery_receipts (
        id, delivery_id, provider_reference_digest, outcome,
        signature_key_id, received_at
    )
    SELECT $1::uuid, id, $3::bytea, $4, $5, $6::timestamptz
    FROM current_delivery
    WHERE state = 'accepted'
    ON CONFLICT (delivery_id, provider_reference_digest, outcome) DO NOTHING
    RETURNING delivery_id
), transitioned AS (
    UPDATE notifications_deliveries AS delivery
    SET state = CASE WHEN $4 = 'delivered' THEN 'delivered' ELSE 'bounced' END,
        outcome_code = CASE
            WHEN $4 = 'delivered' THEN 'DELIVERED'
            WHEN $4 = 'bounced' THEN 'BOUNCED'
            ELSE 'COMPLAINED'
        END,
        row_version = row_version + 1
    FROM current_delivery
    WHERE delivery.id = current_delivery.id
      AND current_delivery.state = 'accepted'
      AND EXISTS (SELECT 1 FROM inserted_receipt)
    RETURNING delivery.id
), accepted_result AS (
    SELECT id, false AS replay FROM transitioned
    UNION ALL
    SELECT current_delivery.id, true
    FROM current_delivery
    WHERE NOT EXISTS (SELECT 1 FROM transitioned)
      AND current_delivery.state = CASE
          WHEN $4 = 'delivered' THEN 'delivered' ELSE 'bounced' END
      AND EXISTS (
          SELECT 1 FROM notifications_delivery_receipts AS receipt
          WHERE receipt.delivery_id = current_delivery.id
            AND receipt.provider_reference_digest = $3::bytea
            AND receipt.outcome = $4
      )
), suppression AS (
    INSERT INTO notifications_suppressed_targets (
        channel, target_digest, reason, created_at
    )
    SELECT current_delivery.channel, $7::bytea,
           CASE WHEN $4 = 'complained' THEN 'complaint' ELSE 'bounce' END,
           $6::timestamptz
    FROM current_delivery
    JOIN accepted_result ON accepted_result.id = current_delivery.id
    WHERE $4 IN ('bounced', 'complained')
    ON CONFLICT (channel, target_digest) DO UPDATE
    SET reason = CASE
            WHEN notifications_suppressed_targets.reason = 'complaint'
              OR EXCLUDED.reason = 'complaint' THEN 'complaint'
            ELSE notifications_suppressed_targets.reason
        END,
        created_at = LEAST(notifications_suppressed_targets.created_at, EXCLUDED.created_at)
)
SELECT id::text, replay FROM accepted_result;
