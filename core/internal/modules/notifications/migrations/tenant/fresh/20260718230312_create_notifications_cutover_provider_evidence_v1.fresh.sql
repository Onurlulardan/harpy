-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_provider_evidence_v1(p_scope_id uuid)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_delivery_sha text;
    v_attempt_sha text;
    v_receipt_sha text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Notifications V1 provider evidence principal denied';
    END IF;
    IF EXISTS (
        SELECT 1 FROM notifications_deliveries AS delivery
        WHERE delivery.state IN ('pending','sending') OR delivery.outcome_unknown
           OR delivery.lease_owner<>'' OR delivery.lease_until IS NOT NULL
    ) OR EXISTS (
        SELECT 1 FROM notifications_effect_receipts AS effect
        WHERE effect.outcome IN ('pending','outcome_unknown')
    ) OR EXISTS (
        SELECT 1 FROM notifications_deliveries AS delivery
        WHERE NOT EXISTS (
            SELECT 1 FROM notifications_delivery_attempts AS attempt
            WHERE attempt.delivery_id=delivery.id
        ) OR (SELECT pg_catalog.count(*)
              FROM notifications_delivery_attempts AS attempt
              WHERE attempt.delivery_id=delivery.id)<>delivery.attempt::bigint
          OR (SELECT pg_catalog.max(attempt.attempt_number)
              FROM notifications_delivery_attempts AS attempt
              WHERE attempt.delivery_id=delivery.id) IS DISTINCT FROM delivery.attempt
    ) OR EXISTS (
        SELECT 1 FROM notifications_deliveries AS delivery
        WHERE delivery.state='accepted' AND NOT EXISTS (
            SELECT 1 FROM notifications_delivery_attempts AS attempt
            WHERE attempt.delivery_id=delivery.id
              AND attempt.response_class='legacy_v1_accepted'
        )
    ) OR EXISTS (
        SELECT 1 FROM notifications_deliveries AS delivery
        WHERE delivery.state IN ('delivered','bounced') AND NOT EXISTS (
            SELECT 1 FROM notifications_delivery_receipts AS receipt
            WHERE receipt.delivery_id=delivery.id
              AND receipt.outcome=delivery.state
              AND receipt.signature_key_id NOT LIKE 'legacy%'
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 provider receipt evidence is incomplete';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
        pg_catalog.string_agg(pg_catalog.to_jsonb(delivery)::text,',' ORDER BY delivery.id),
        ''),'UTF8')),'hex') INTO v_delivery_sha FROM notifications_deliveries AS delivery;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
        pg_catalog.string_agg(pg_catalog.to_jsonb(attempt)::text,',' ORDER BY attempt.id),
        ''),'UTF8')),'hex') INTO v_attempt_sha
    FROM notifications_delivery_attempts AS attempt;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(COALESCE(
        pg_catalog.string_agg(pg_catalog.to_jsonb(receipt)::text,',' ORDER BY receipt.id),
        ''),'UTF8')),'hex') INTO v_receipt_sha
    FROM notifications_delivery_receipts AS receipt;
    RETURN pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'forgecrud/notifications/v1-provider-evidence/v1:'||p_scope_id::text||':'||
        v_delivery_sha||':'||v_attempt_sha||':'||v_receipt_sha,
        'UTF8'
    )),'hex');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_provider_evidence_v1(uuid) FROM PUBLIC;
