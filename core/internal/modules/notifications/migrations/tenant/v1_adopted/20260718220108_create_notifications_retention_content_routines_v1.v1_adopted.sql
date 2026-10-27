-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_retention_redact_content_v1(
    p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE v_count bigint;
BEGIN
WITH candidates AS MATERIALIZED (
    SELECT id FROM notifications_messages
    WHERE pg_catalog.isfinite(p_now) AND p_limit BETWEEN 1 AND 500
      AND purged_at IS NULL AND (
        (dismissed_at IS NOT NULL AND dismissed_at <= p_now - interval '30 days')
        OR (dismissed_at IS NULL AND read_at IS NOT NULL
            AND read_at <= p_now - interval '90 days')
        OR (dismissed_at IS NULL AND read_at IS NULL
            AND created_at <= p_now - interval '180 days')
      )
    ORDER BY COALESCE(dismissed_at, read_at, created_at), id
    FOR UPDATE SKIP LOCKED LIMIT p_limit
), redacted AS (
    UPDATE notifications_messages AS message
    SET subject = '', body = '', purged_at = p_now,
        row_version = message.row_version + 1
    FROM candidates WHERE message.id = candidates.id RETURNING message.id
)
SELECT pg_catalog.count(*) INTO v_count FROM redacted;
RETURN v_count;
END;
$$;

CREATE FUNCTION notifications_retention_purge_attempts_v1(
    p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE v_count bigint;
BEGIN
WITH candidates AS MATERIALIZED (
    SELECT id FROM notifications_delivery_attempts
    WHERE pg_catalog.isfinite(p_now) AND p_limit BETWEEN 1 AND 500
      AND finished_at <= p_now - interval '90 days'
    ORDER BY finished_at, id FOR UPDATE SKIP LOCKED LIMIT p_limit
), purged AS (
    DELETE FROM notifications_delivery_attempts AS attempt
    USING candidates WHERE attempt.id = candidates.id RETURNING attempt.id
)
SELECT pg_catalog.count(*) INTO v_count FROM purged;
RETURN v_count;
END;
$$;

CREATE FUNCTION notifications_retention_purge_receipts_v1(
    p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE v_count bigint;
BEGIN
WITH candidates AS MATERIALIZED (
    SELECT id FROM notifications_delivery_receipts
    WHERE pg_catalog.isfinite(p_now) AND p_limit BETWEEN 1 AND 500
      AND received_at <= p_now - interval '90 days'
    ORDER BY received_at, id FOR UPDATE SKIP LOCKED LIMIT p_limit
), purged AS (
    DELETE FROM notifications_delivery_receipts AS receipt
    USING candidates WHERE receipt.id = candidates.id RETURNING receipt.id
)
SELECT pg_catalog.count(*) INTO v_count FROM purged;
RETURN v_count;
END;
$$;

CREATE FUNCTION notifications_retention_purge_effects_v1(
    p_now timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $$
DECLARE v_count bigint;
BEGIN
WITH candidates AS MATERIALIZED (
    SELECT producer, effect_key_hash, recipient_id
    FROM notifications_effect_receipts
    WHERE pg_catalog.isfinite(p_now) AND p_limit BETWEEN 1 AND 500
      AND retain_until <= p_now
    ORDER BY retain_until, producer, effect_key_hash, recipient_id
    FOR UPDATE SKIP LOCKED LIMIT p_limit
), purged AS (
    DELETE FROM notifications_effect_receipts AS receipt USING candidates
    WHERE receipt.producer = candidates.producer
      AND receipt.effect_key_hash = candidates.effect_key_hash
      AND receipt.recipient_id = candidates.recipient_id
    RETURNING receipt.message_id
)
SELECT pg_catalog.count(*) INTO v_count FROM purged;
RETURN v_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_retention_redact_content_v1(timestamptz, integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION notifications_retention_purge_attempts_v1(timestamptz, integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION notifications_retention_purge_receipts_v1(timestamptz, integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION notifications_retention_purge_effects_v1(timestamptz, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION notifications_retention_redact_content_v1(timestamptz, integer) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_attempts_v1(timestamptz, integer) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_receipts_v1(timestamptz, integer) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_effects_v1(timestamptz, integer) TO harpy_runtime;
