-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_retention_purge_hints_v1(
    p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE v_count bigint;
BEGIN
WITH candidates AS MATERIALIZED (
    SELECT id FROM notifications_realtime_hints
    WHERE pg_catalog.isfinite(p_now) AND p_limit BETWEEN 1 AND 500
      AND created_at <= p_now - interval '24 hours'
    ORDER BY created_at, id FOR UPDATE SKIP LOCKED LIMIT p_limit
), purged AS (
    DELETE FROM notifications_realtime_hints AS hint
    USING candidates WHERE hint.id = candidates.id RETURNING hint.id
)
SELECT pg_catalog.count(*) INTO v_count FROM purged;
RETURN v_count;
END;
$$;

CREATE FUNCTION notifications_retention_purge_deliveries_v1(
    p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE v_count bigint;
BEGIN
WITH candidates AS MATERIALIZED (
    SELECT delivery.id
    FROM notifications_deliveries AS delivery
    JOIN notifications_messages AS message ON message.id = delivery.message_id
    WHERE pg_catalog.isfinite(p_now) AND p_limit BETWEEN 1 AND 500
      AND message.purged_at <= p_now - interval '365 days'
      AND NOT EXISTS (
          SELECT 1 FROM notifications_effect_receipts AS effect
          WHERE effect.message_id = message.id
      )
      AND NOT EXISTS (
          SELECT 1 FROM notifications_delivery_attempts AS attempt
          WHERE attempt.delivery_id = delivery.id
      )
      AND NOT EXISTS (
          SELECT 1 FROM notifications_delivery_receipts AS receipt
          WHERE receipt.delivery_id = delivery.id
      )
    ORDER BY message.purged_at, delivery.id
    FOR UPDATE OF delivery SKIP LOCKED LIMIT p_limit
), purged AS (
    DELETE FROM notifications_deliveries AS delivery
    USING candidates WHERE delivery.id = candidates.id RETURNING delivery.id
)
SELECT pg_catalog.count(*) INTO v_count FROM purged;
RETURN v_count;
END;
$$;

CREATE FUNCTION notifications_retention_purge_messages_v1(
    p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE v_count bigint;
BEGIN
WITH candidates AS MATERIALIZED (
    SELECT message.id FROM notifications_messages AS message
    WHERE pg_catalog.isfinite(p_now) AND p_limit BETWEEN 1 AND 500
      AND message.purged_at <= p_now - interval '365 days'
      AND NOT EXISTS (
          SELECT 1 FROM notifications_effect_receipts AS effect
          WHERE effect.message_id = message.id
      )
      AND NOT EXISTS (
          SELECT 1 FROM notifications_deliveries AS delivery
          WHERE delivery.message_id = message.id
      )
      AND NOT EXISTS (
          SELECT 1 FROM notifications_realtime_hints AS hint
          WHERE hint.message_id = message.id
      )
    ORDER BY message.purged_at, message.id
    FOR UPDATE OF message SKIP LOCKED LIMIT p_limit
), purged AS (
    DELETE FROM notifications_messages AS message
    USING candidates WHERE message.id = candidates.id RETURNING message.id
)
SELECT pg_catalog.count(*) INTO v_count FROM purged;
RETURN v_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_retention_purge_hints_v1(timestamptz, integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION notifications_retention_purge_deliveries_v1(timestamptz, integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION notifications_retention_purge_messages_v1(timestamptz, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_hints_v1(timestamptz, integer) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_deliveries_v1(timestamptz, integer) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_messages_v1(timestamptz, integer) TO forgecrud_runtime;
