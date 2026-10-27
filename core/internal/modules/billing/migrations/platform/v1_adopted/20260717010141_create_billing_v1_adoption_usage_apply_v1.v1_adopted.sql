-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_opening_v1(
    p_operation_id uuid, p_run_id uuid, p_record_kind text, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_definition platform.billing_entitlement_definitions%ROWTYPE;
    v_period_id uuid;
    v_entry_id uuid;
    v_value bigint;
    v_period_start timestamptz;
    v_period_end timestamptz;
    v_evidence text;
    v_source_key text;
    v_category text;
    v_meter_kind text;
    v_reset_policy text;
    v_owner_fact_id uuid;
    v_sha bytea;
BEGIN
    v_period_id := (p_record->>'meter_period_id')::uuid;
    v_entry_id := NULLIF(p_record->>'entry_id', '')::uuid;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'tenant' AND run.state = 'planned' AND run.eligible;
    IF (p_record->>'tenant_id')::uuid IS DISTINCT FROM v_run.scope_id
       OR p_record_kind NOT IN ('usage_opening', 'gauge_opening') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 opening scope is invalid';
    END IF;
    IF p_record_kind = 'usage_opening' THEN
        v_value := (p_record->>'count')::bigint;
        v_period_start := (p_record->>'period_start')::timestamptz;
        v_period_end := (p_record->>'period_end')::timestamptz;
        v_evidence := p_record->>'flush_evidence_sha256';
        v_source_key := p_record->>'source_id';
        v_category := 'submission_usage';
        v_meter_kind := 'counter';
        v_reset_policy := 'calendar_month_utc';
    ELSE
        v_value := (p_record->>'value')::bigint;
        v_period_start := (p_record->>'measured_at')::timestamptz;
        v_period_end := '9999-12-31 23:59:59.999999+00'::timestamptz;
        v_evidence := p_record->>'evidence_sha256';
        v_owner_fact_id := (p_record->>'owner_fact_id')::uuid;
        v_source_key := p_record->>'owner_fact_id';
        v_category := 'gauge_openings';
        v_meter_kind := 'gauge';
        v_reset_policy := 'never';
    END IF;
    SELECT definition.* INTO STRICT v_definition
    FROM platform.billing_entitlement_definitions AS definition
    WHERE definition.entitlement_key = p_record->>'entitlement_key';
    INSERT INTO platform.billing_meter_periods (
        id, authority_id, tenant_id, entitlement_definition_id,
        entitlement_key, meter_kind, reset_policy, period_start,
        period_end, state, closed_at, revision
    ) VALUES (
        v_period_id, v_run.authority_id, v_run.scope_id, v_definition.id,
        p_record->>'entitlement_key', v_meter_kind, v_reset_policy,
        v_period_start, v_period_end, 'open', NULL, 1
    );
    INSERT INTO platform.billing_usage_totals (
        authority_id, tenant_id, meter_period_id, entitlement_key,
        meter_kind, committed_total, reserved_total, gauge_total,
        revision, updated_at
    ) VALUES (
        v_run.authority_id, v_run.scope_id, v_period_id,
        p_record->>'entitlement_key', v_meter_kind,
        CASE WHEN v_meter_kind = 'counter' THEN v_value ELSE 0 END,
        0, CASE WHEN v_meter_kind = 'gauge' THEN v_value ELSE 0 END,
        1, v_run.snapshot_at
    );
    IF v_value > 0 THEN
        INSERT INTO platform.billing_usage_entries (
            id, authority_id, tenant_id, decision_id, entitlement_key,
            meter_period_id, reservation_id, entry_kind, amount_delta,
            committed_after, reserved_after, gauge_after,
            source_operation_id, owner_fact_id, evidence_sha256,
            occurred_at, created_at
        ) VALUES (
            v_entry_id, v_run.authority_id, v_run.scope_id, NULL,
            p_record->>'entitlement_key', v_period_id, NULL, 'opening', v_value,
            CASE WHEN v_meter_kind = 'counter' THEN v_value ELSE 0 END,
            0, CASE WHEN v_meter_kind = 'gauge' THEN v_value ELSE 0 END,
            NULL, NULL, pg_catalog.decode(v_evidence, 'hex'),
            v_period_start, v_run.snapshot_at
        );
    END IF;
    INSERT INTO platform.billing_v1_adoption_opening_evidence (
        operation_id, run_id, scope_kind, scope_id, authority_id,
        evidence_kind, source_key, tenant_id, entitlement_key,
        meter_period_id, entry_id, owner_fact_id, period_start,
        period_end, opening_value, evidence_sha256, recorded_at
    ) VALUES (
        p_operation_id, p_run_id, v_run.scope_kind, v_run.scope_id,
        v_run.authority_id,
        CASE p_record_kind WHEN 'usage_opening' THEN 'submission_usage'
            ELSE 'owner_gauge' END,
        v_source_key, v_run.scope_id, p_record->>'entitlement_key',
        v_period_id, v_entry_id, v_owner_fact_id, v_period_start,
        v_period_end, v_value, pg_catalog.decode(v_evidence, 'hex'),
        v_run.snapshot_at
    );
    v_sha := platform.billing_v1_adoption_tenant_fingerprint_v1(
        p_record_kind, v_period_id
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, v_category, v_source_key,
        p_record_kind, v_period_id, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_opening_v1(
    uuid, uuid, text, jsonb
) FROM PUBLIC;
-- +goose StatementEnd
