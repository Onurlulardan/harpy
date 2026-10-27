-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_redemption_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_registry bigint;
    v_id uuid;
    v_sha bytea;
BEGIN
    v_id := (p_record->>'id')::uuid;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'tenant' AND run.state = 'planned' AND run.eligible;
    IF (p_record->>'tenant_id')::uuid IS DISTINCT FROM v_run.scope_id THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 redemption tenant is invalid';
    END IF;
    IF p_record->>'applied_kind' = 'fixed' THEN
        SELECT exponent.registry_version INTO STRICT v_registry
        FROM platform.billing_currency_exponents AS exponent
        WHERE exponent.code = (p_record->>'applied_currency')::character(3)
          AND exponent.financial_money_eligible
        ORDER BY exponent.registry_version DESC LIMIT 1;
    END IF;
    INSERT INTO platform.billing_discount_redemptions (
        id, authority_id, tenant_id, subscription_id, discount_version_id,
        offering_version_id, price_version_id, source_operation_id,
        applied_kind, applied_fixed_minor_units, applied_currency,
        applied_currency_registry_version, applied_basis_points,
        applied_priority, applied_duration, applied_repeating_cycles,
        policy_sha256, redeemed_at, relationship_ended_at
    ) VALUES (
        v_id, v_run.authority_id, v_run.scope_id,
        (p_record->>'subscription_id')::uuid,
        (p_record->>'discount_version_id')::uuid,
        (p_record->>'offering_version_id')::uuid,
        (p_record->>'price_version_id')::uuid,
        (p_record->>'source_operation_id')::uuid,
        p_record->>'applied_kind',
        NULLIF(p_record->>'applied_fixed_minor_units', '')::bigint,
        NULLIF(p_record->>'applied_currency', '')::character(3), v_registry,
        NULLIF(p_record->>'applied_basis_points', '')::integer,
        (p_record->>'applied_priority')::integer,
        p_record->>'applied_duration',
        NULLIF(p_record->>'applied_repeating_cycles', '')::bigint,
        pg_catalog.decode(p_record->>'policy_sha256', 'hex'),
        (p_record->>'redeemed_at')::timestamptz,
        NULLIF(p_record->>'relationship_ended_at', '')::timestamptz
    );
    v_sha := platform.billing_v1_adoption_tenant_fingerprint_v1(
        'redemption', v_id
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'subscription_coupons',
        p_record->>'source_id', 'redemption', v_id, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_redemption_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;
-- +goose StatementEnd
