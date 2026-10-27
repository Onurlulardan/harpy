-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_purge_entitlement_retention_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_now timestamptz,
    p_limit integer
)
RETURNS TABLE (records_purged bigint, has_remaining boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    decision_count bigint;
    total_count bigint;
    period_count bigint;
BEGIN
    decision_count := 0;
    total_count := 0;
    period_count := 0;
    IF NOT platform.billing_retention_delete_allowed_v1(
        p_authority_id, p_tenant_id
    ) OR p_limit NOT BETWEEN 1 AND 1000 OR NOT pg_catalog.isfinite(p_now) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing entitlement retention context is invalid';
    END IF;
    WITH candidates AS (
        SELECT decision.id
        FROM platform.billing_entitlement_decisions AS decision
        WHERE decision.tenant_id = p_tenant_id
          AND decision.retain_until <= p_now
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_reservations AS reservation
              WHERE reservation.tenant_id = p_tenant_id
                AND reservation.decision_id = decision.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_entries AS entry
              WHERE entry.tenant_id = p_tenant_id
                AND entry.decision_id = decision.id
          )
        ORDER BY decision.retain_until, decision.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ), removed AS (
        DELETE FROM platform.billing_entitlement_decisions AS decision
        USING candidates
        WHERE decision.tenant_id = p_tenant_id AND decision.id = candidates.id
        RETURNING 1
    ) SELECT pg_catalog.count(*) INTO decision_count FROM removed;

    WITH candidates AS (
        SELECT total.meter_period_id
        FROM platform.billing_usage_totals AS total
        JOIN platform.billing_meter_periods AS period
          ON period.tenant_id = total.tenant_id
         AND period.id = total.meter_period_id
        WHERE total.tenant_id = p_tenant_id AND period.state = 'closed'
          AND period.closed_at <= p_now - INTERVAL '7 years'
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_reservations AS reservation
              WHERE reservation.tenant_id = p_tenant_id
                AND reservation.meter_period_id = total.meter_period_id
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_entries AS entry
              WHERE entry.tenant_id = p_tenant_id
                AND entry.meter_period_id = total.meter_period_id
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_entitlement_decisions AS decision
              WHERE decision.tenant_id = p_tenant_id
                AND decision.meter_period_id = total.meter_period_id
          )
        ORDER BY period.closed_at, total.meter_period_id
        FOR UPDATE OF total SKIP LOCKED LIMIT p_limit
    ), removed AS (
        DELETE FROM platform.billing_usage_totals AS total
        USING candidates
        WHERE total.tenant_id = p_tenant_id
          AND total.meter_period_id = candidates.meter_period_id
        RETURNING 1
    ) SELECT pg_catalog.count(*) INTO total_count FROM removed;

    WITH candidates AS (
        SELECT period.id
        FROM platform.billing_meter_periods AS period
        WHERE period.tenant_id = p_tenant_id AND period.state = 'closed'
          AND period.closed_at <= p_now - INTERVAL '7 years'
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_totals AS total
              WHERE total.tenant_id = p_tenant_id
                AND total.meter_period_id = period.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_reservations AS reservation
              WHERE reservation.tenant_id = p_tenant_id
                AND reservation.meter_period_id = period.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_usage_entries AS entry
              WHERE entry.tenant_id = p_tenant_id
                AND entry.meter_period_id = period.id
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.billing_entitlement_decisions AS decision
              WHERE decision.tenant_id = p_tenant_id
                AND decision.meter_period_id = period.id
          )
        ORDER BY period.closed_at, period.id
        FOR UPDATE SKIP LOCKED LIMIT p_limit
    ), removed AS (
        DELETE FROM platform.billing_meter_periods AS period
        USING candidates
        WHERE period.tenant_id = p_tenant_id AND period.id = candidates.id
        RETURNING 1
    ) SELECT pg_catalog.count(*) INTO period_count FROM removed;

    RETURN QUERY SELECT decision_count + total_count + period_count,
        decision_count = p_limit OR total_count = p_limit
        OR period_count = p_limit;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_purge_entitlement_retention_v1(
    uuid,uuid,timestamptz,integer
) FROM PUBLIC;
-- +goose StatementEnd
