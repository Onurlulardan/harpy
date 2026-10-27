-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_purge_raw_retention_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_now timestamptz,
    p_limit integer
)
RETURNS TABLE (
    raw_bodies_redacted bigint,
    raw_records_purged bigint,
    has_remaining boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    body_count bigint;
    entry_count bigint;
    reservation_count bigint;
    receipt_count bigint;
BEGIN
    body_count := 0;
    entry_count := 0;
    reservation_count := 0;
    receipt_count := 0;
    IF NOT platform.billing_retention_delete_allowed_v1(
        p_authority_id, p_tenant_id
    ) OR p_limit NOT BETWEEN 1 AND 1000 OR NOT pg_catalog.isfinite(p_now) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing raw retention context is invalid';
    END IF;
    WITH candidates AS (
        SELECT receipt.id
        FROM platform.billing_webhook_receipts AS receipt
        WHERE receipt.tenant_id = p_tenant_id
          AND receipt.raw_body_expires_at <= p_now
          AND receipt.encrypted_body IS NOT NULL
        ORDER BY receipt.raw_body_expires_at, receipt.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ), changed AS (
        UPDATE platform.billing_webhook_receipts AS receipt
           SET encrypted_body = NULL, wrapped_dek = NULL,
               algorithm = NULL, nonce = NULL, revision = receipt.revision + 1
          FROM candidates
         WHERE receipt.tenant_id = p_tenant_id AND receipt.id = candidates.id
         RETURNING 1
    ) SELECT pg_catalog.count(*) INTO body_count FROM changed;

    WITH candidates AS (
        SELECT entry.id
        FROM platform.billing_usage_entries AS entry
        WHERE entry.tenant_id = p_tenant_id
          AND entry.occurred_at <= p_now - INTERVAL '400 days'
        ORDER BY entry.occurred_at, entry.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ), removed AS (
        DELETE FROM platform.billing_usage_entries AS entry
        USING candidates
        WHERE entry.tenant_id = p_tenant_id AND entry.id = candidates.id
        RETURNING 1
    ) SELECT pg_catalog.count(*) INTO entry_count FROM removed;

    WITH candidates AS (
        SELECT reservation.id
        FROM platform.billing_usage_reservations AS reservation
        WHERE reservation.tenant_id = p_tenant_id
          AND reservation.state IN ('committed', 'released')
          AND reservation.resolved_at <= p_now - INTERVAL '400 days'
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_entries AS entry
              WHERE entry.tenant_id = p_tenant_id
                AND entry.reservation_id = reservation.id
          )
        ORDER BY reservation.resolved_at, reservation.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ), removed AS (
        DELETE FROM platform.billing_usage_reservations AS reservation
        USING candidates
        WHERE reservation.tenant_id = p_tenant_id
          AND reservation.id = candidates.id
        RETURNING 1
    ) SELECT pg_catalog.count(*) INTO reservation_count FROM removed;

    WITH candidates AS (
        SELECT receipt.id
        FROM platform.billing_webhook_receipts AS receipt
        WHERE receipt.tenant_id = p_tenant_id
          AND receipt.retain_until <= p_now
          AND receipt.processor_state IN (
              'applied', 'reconcile_required', 'rejected'
          )
        ORDER BY receipt.retain_until, receipt.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ), removed AS (
        DELETE FROM platform.billing_webhook_receipts AS receipt
        USING candidates
        WHERE receipt.tenant_id = p_tenant_id AND receipt.id = candidates.id
        RETURNING 1
    ) SELECT pg_catalog.count(*) INTO receipt_count FROM removed;

    RETURN QUERY SELECT body_count,
        entry_count + reservation_count + receipt_count,
        body_count = p_limit OR entry_count = p_limit
        OR reservation_count = p_limit OR receipt_count = p_limit;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_purge_raw_retention_v1(
    uuid,uuid,timestamptz,integer
) FROM PUBLIC;
-- +goose StatementEnd
