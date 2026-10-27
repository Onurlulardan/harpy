-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_discount_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_offering jsonb;
    v_registry bigint;
    v_sha bytea;
BEGIN
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'global' AND run.state = 'planned' AND run.eligible;
    IF p_record->>'discount_kind' = 'fixed' THEN
        SELECT exponent.registry_version INTO STRICT v_registry
        FROM platform.billing_currency_exponents AS exponent
        WHERE exponent.authority_id = v_run.authority_id
          AND exponent.code = (p_record->>'currency')::character(3)
          AND exponent.financial_money_eligible
        ORDER BY exponent.registry_version DESC LIMIT 1;
    END IF;
    INSERT INTO platform.billing_discounts (
        id, authority_id, normalized_code, lifecycle, created_at, revoked_at, revision,
        legacy_created_by
    ) VALUES (
        (p_record->>'id')::uuid, v_run.authority_id,
        p_record->>'normalized_code',
        CASE WHEN p_record->>'revoked_at' IS NULL THEN 'active' ELSE 'revoked' END,
        (p_record->>'created_at')::timestamptz,
        NULLIF(p_record->>'revoked_at', '')::timestamptz, 1,
        (p_record->>'created_by')::uuid
    );
    INSERT INTO platform.billing_discount_versions (
        id, authority_id, discount_id, version, discount_kind,
        fixed_minor_units, currency, currency_registry_version,
        currency_eligible, basis_points, duration, repeating_cycles,
        priority, global_use_limit, per_tenant_use_limit, all_offerings,
        expected_offering_count, publication_operation_id,
        valid_start, valid_end, published_at,
        presentation_name, presentation_description, available
    ) VALUES (
        (p_record->>'version_id')::uuid, v_run.authority_id,
        (p_record->>'id')::uuid, 1, p_record->>'discount_kind',
        NULLIF(p_record->>'fixed_minor_units', '')::bigint,
        NULLIF(p_record->>'currency', '')::character(3), v_registry,
        CASE WHEN p_record->>'discount_kind' = 'fixed' THEN true ELSE NULL END,
        NULLIF(p_record->>'basis_points', '')::integer,
        p_record->>'duration', NULLIF(p_record->>'repeating_cycles', '')::bigint,
        (p_record->>'priority')::integer,
        NULLIF(p_record->>'global_use_limit', '')::bigint,
        NULLIF(p_record->>'per_tenant_use_limit', '')::bigint,
        (p_record->>'all_offerings')::boolean,
        pg_catalog.jsonb_array_length(p_record->'offering_version_ids'),
        (p_record->>'publication_operation_id')::uuid,
        (p_record->>'valid_start')::timestamptz,
        NULLIF(p_record->>'valid_end', '')::timestamptz,
        (p_record->>'updated_at')::timestamptz,
        p_record->>'name', p_record->>'description',
        (p_record->>'is_active')::boolean
    );
    FOR v_offering IN
        SELECT value
        FROM pg_catalog.jsonb_array_elements(p_record->'offering_version_ids')
    LOOP
        INSERT INTO platform.billing_discount_offerings (
            authority_id, discount_version_id, offering_version_id, created_at
        ) VALUES (
            v_run.authority_id, (p_record->>'version_id')::uuid,
            (v_offering #>> '{}')::uuid, (p_record->>'created_at')::timestamptz
        );
    END LOOP;
    v_sha := platform.billing_v1_adoption_catalog_fingerprint_v1(
        'discount', (p_record->>'id')::uuid
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'coupons', p_record->>'source_id',
        'discount', (p_record->>'id')::uuid, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_discount_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;
-- +goose StatementEnd
