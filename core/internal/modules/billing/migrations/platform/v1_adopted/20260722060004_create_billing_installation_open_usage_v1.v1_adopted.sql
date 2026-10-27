-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_open_usage_v1(
    p_operation_id uuid, p_tenant_id uuid, p_manifest_sha256 text
)
RETURNS TABLE (meter_count integer, replay boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_assignment platform.billing_installation_base_assignments%ROWTYPE;
    v_opening platform.billing_installation_usage_openings%ROWTYPE;
    v_count integer;
    v_total_count integer;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='billing installation usage input is invalid';
    END IF;
    SELECT * INTO v_assignment
    FROM platform.billing_installation_base_assignments AS assignment
    WHERE assignment.operation_id=p_operation_id
      AND assignment.tenant_id=p_tenant_id
      AND assignment.manifest_sha256=p_manifest_sha256 FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='billing installation assignment is unavailable';
    END IF;
    SELECT * INTO v_opening
    FROM platform.billing_installation_usage_openings AS opening
    WHERE opening.operation_id=p_operation_id;
    IF FOUND THEN
        IF v_opening.tenant_id IS DISTINCT FROM p_tenant_id THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='billing installation usage replay conflicts';
        END IF;
        RETURN QUERY SELECT v_opening.meter_count,true;
        RETURN;
    END IF;
    INSERT INTO platform.billing_meter_periods (
        id,authority_id,tenant_id,entitlement_definition_id,entitlement_key,
        meter_kind,reset_policy,period_start,period_end,state,closed_at,revision
    ) SELECT pg_catalog.uuidv7(),v_assignment.authority_id,p_tenant_id,
        definition.id,definition.entitlement_key,definition.meter_kind,
        definition.reset_policy,v_assignment.offering_effective_at,
        CASE definition.reset_policy
          WHEN 'calendar_month_utc' THEN pg_catalog.date_trunc(
              'month',v_assignment.offering_effective_at AT TIME ZONE 'UTC'
          ) AT TIME ZONE 'UTC' + interval '1 month'
          WHEN 'billing_period' THEN v_assignment.offering_effective_at + interval '1 month'
          ELSE '9999-12-31T23:59:59Z'::timestamptz
        END,'open',NULL,1
      FROM platform.billing_offering_entitlements AS entitlement
      JOIN platform.billing_entitlement_definitions AS definition
        ON definition.id=entitlement.entitlement_definition_id
      WHERE entitlement.offering_version_id=v_assignment.offering_version_id
        AND definition.meter_kind IS NOT NULL
      ORDER BY definition.entitlement_key;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    INSERT INTO platform.billing_usage_totals (
        authority_id,tenant_id,meter_period_id,entitlement_key,meter_kind,
        committed_total,reserved_total,gauge_total,revision,updated_at
    ) SELECT period.authority_id,period.tenant_id,period.id,
        period.entitlement_key,period.meter_kind,0,0,0,1,v_now
      FROM platform.billing_meter_periods AS period
      WHERE period.tenant_id=p_tenant_id
        AND period.period_start=v_assignment.offering_effective_at;
    GET DIAGNOSTICS v_total_count = ROW_COUNT;
    IF v_count IS DISTINCT FROM v_total_count OR v_count < 1 THEN
        RAISE EXCEPTION USING ERRCODE='23503',
            MESSAGE='billing installation usage opening is incomplete';
    END IF;
    INSERT INTO platform.billing_installation_usage_openings (
        operation_id,tenant_id,meter_count,opened_at
    ) VALUES (p_operation_id,p_tenant_id,v_count,v_now);
    RETURN QUERY SELECT v_count,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_installation_open_usage_v1(
    uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_installation_open_usage_v1(
    uuid,uuid,text
) TO harpy_runtime;
