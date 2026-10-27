SELECT EXISTS (
    SELECT 1 FROM notifications_messages WHERE purged_at IS NULL AND (
        (dismissed_at IS NOT NULL AND dismissed_at <= $1::timestamptz - interval '30 days')
        OR (dismissed_at IS NULL AND read_at IS NOT NULL
            AND read_at <= $1::timestamptz - interval '90 days')
        OR (dismissed_at IS NULL AND read_at IS NULL
            AND created_at <= $1::timestamptz - interval '180 days')
    )
) OR EXISTS (
    SELECT 1 FROM notifications_delivery_attempts
    WHERE finished_at <= $1::timestamptz - interval '90 days'
) OR EXISTS (
    SELECT 1 FROM notifications_delivery_receipts
    WHERE received_at <= $1::timestamptz - interval '90 days'
) OR EXISTS (
    SELECT 1 FROM notifications_effect_receipts WHERE retain_until <= $1::timestamptz
) OR EXISTS (
    SELECT 1 FROM notifications_realtime_hints
    WHERE created_at <= $1::timestamptz - interval '24 hours'
) OR EXISTS (
    SELECT 1 FROM notifications_deliveries AS delivery
    JOIN notifications_messages AS message ON message.id = delivery.message_id
    WHERE message.purged_at <= $1::timestamptz - interval '365 days'
      AND NOT EXISTS (SELECT 1 FROM notifications_effect_receipts AS effect
                      WHERE effect.message_id = message.id)
      AND NOT EXISTS (SELECT 1 FROM notifications_delivery_attempts AS attempt
                      WHERE attempt.delivery_id = delivery.id)
      AND NOT EXISTS (SELECT 1 FROM notifications_delivery_receipts AS receipt
                      WHERE receipt.delivery_id = delivery.id)
) OR EXISTS (
    SELECT 1 FROM notifications_messages AS message
    WHERE message.purged_at <= $1::timestamptz - interval '365 days'
      AND NOT EXISTS (SELECT 1 FROM notifications_effect_receipts AS effect
                      WHERE effect.message_id = message.id)
      AND NOT EXISTS (SELECT 1 FROM notifications_deliveries AS delivery
                      WHERE delivery.message_id = message.id)
      AND NOT EXISTS (SELECT 1 FROM notifications_realtime_hints AS hint
                      WHERE hint.message_id = message.id)
);
