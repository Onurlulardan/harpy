-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_fingerprint_v1(p_record_kind text, p_target_id uuid)
RETURNS bytea LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF p_record_kind IN ('offering', 'campaign', 'discount', 'provider_account') THEN
        RETURN platform.billing_v1_adoption_catalog_fingerprint_v1(p_record_kind, p_target_id);
    END IF;
    RETURN platform.billing_v1_adoption_tenant_fingerprint_v1(p_record_kind, p_target_id);
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_v1_adoption_fingerprint_v1(text, uuid) FROM PUBLIC;
CREATE FUNCTION platform.billing_verify_v1_adoption_v1(p_operation_id uuid, p_run_id uuid)
RETURNS TABLE (
    source_count bigint, target_count bigint, source_sha256 text, target_sha256 text,
    money_difference bigint, period_difference bigint, coupon_difference bigint,
    usage_difference bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_baseline bytea;
    v_target bytea;
    v_source_count bigint;
    v_target_count bigint;
    v_money bigint;
    v_period bigint;
    v_coupon bigint;
    v_usage bigint;
    v_now timestamptz;
BEGIN
    v_money := 0; v_period := 0; v_coupon := 0; v_usage := 0;
    v_now := pg_catalog.clock_timestamp();
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
    FOR UPDATE;
    IF v_run.state NOT IN ('applied', 'verified', 'rollback_rehearsed') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing V1 adoption run is not verifiable';
    END IF; SELECT COALESCE(pg_catalog.sum(category.source_count), 0)
    INTO v_source_count
    FROM platform.billing_v1_adoption_categories AS category
    WHERE category.operation_id = p_operation_id AND category.run_id = p_run_id;
    SELECT pg_catalog.sha256(pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
               mapping.record_kind || ':' || mapping.source_key || ':' ||
               pg_catalog.encode(mapping.record_sha256, 'hex'), '|'
               ORDER BY mapping.record_kind, mapping.source_key
           ), ''), 'UTF8'))
    INTO v_baseline
    FROM platform.billing_v1_adoption_mappings AS mapping
    WHERE mapping.operation_id = p_operation_id AND mapping.run_id = p_run_id;
    SELECT pg_catalog.count(current_sha),
           pg_catalog.sha256(pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
               record_kind || ':' || source_key || ':' ||
               pg_catalog.encode(current_sha, 'hex'), '|'
               ORDER BY record_kind, source_key
           ) FILTER (WHERE current_sha IS NOT NULL), ''), 'UTF8'))
    INTO v_target_count, v_target
    FROM (
        SELECT mapping.record_kind, mapping.source_key,
            platform.billing_v1_adoption_fingerprint_v1(
                mapping.record_kind, mapping.target_id
            ) AS current_sha
        FROM platform.billing_v1_adoption_mappings AS mapping
        WHERE mapping.operation_id = p_operation_id AND mapping.run_id = p_run_id
    ) AS current_rows;
    IF v_run.scope_kind = 'global' THEN
        SELECT COALESCE(pg_catalog.sum(pg_catalog.abs(
            (discount->>'source_used_count')::bigint - (
                SELECT pg_catalog.count(*)
                FROM platform.billing_discount_redemptions AS redemption
                WHERE redemption.discount_version_id =
                    (discount->>'version_id')::uuid
            )
        )), 0) INTO v_coupon
        FROM pg_catalog.jsonb_array_elements(
            v_run.plan_document->'discounts'
        ) AS discount;
    ELSE
        SELECT COALESCE(pg_catalog.sum(pg_catalog.abs(
            (subscription->>'effective_minor_units')::bigint -
            COALESCE(effective.effective_minor_units, -1)
        )), 0) INTO v_money
        FROM pg_catalog.jsonb_array_elements(
            v_run.plan_document->'subscriptions'
        ) AS subscription
        LEFT JOIN platform.billing_effective_subscription_price_v1(
            v_run.authority_id, v_run.scope_id,
            (subscription->>'id')::uuid, v_run.snapshot_at
        ) AS effective ON true;
        SELECT pg_catalog.count(*) INTO v_period
        FROM pg_catalog.jsonb_array_elements(
            v_run.plan_document->'subscriptions'
        ) AS expected
        LEFT JOIN platform.billing_subscriptions AS actual
          ON actual.id = (expected->>'id')::uuid
        WHERE actual.id IS NULL OR actual.state IS DISTINCT FROM expected->>'state'
           OR actual.period_start IS DISTINCT FROM (expected->>'period_start')::timestamptz
           OR actual.period_end IS DISTINCT FROM (expected->>'period_end')::timestamptz
           OR actual.trial_ends_at IS DISTINCT FROM
                NULLIF(expected->>'trial_ends_at', '')::timestamptz
           OR actual.grace_ends_at IS DISTINCT FROM
                NULLIF(expected->>'grace_ends_at', '')::timestamptz
           OR actual.cancel_at IS DISTINCT FROM
                NULLIF(expected->>'cancel_at', '')::timestamptz;
        SELECT pg_catalog.count(*) INTO v_coupon
        FROM pg_catalog.jsonb_array_elements(
            v_run.plan_document->'redemptions'
        ) AS expected
        LEFT JOIN platform.billing_discount_redemptions AS actual
          ON actual.id = (expected->>'id')::uuid
        WHERE actual.id IS NULL
           OR actual.policy_sha256 IS DISTINCT FROM
                pg_catalog.decode(expected->>'policy_sha256', 'hex')
           OR actual.applied_kind IS DISTINCT FROM expected->>'applied_kind'
           OR actual.applied_fixed_minor_units IS DISTINCT FROM
                NULLIF(expected->>'applied_fixed_minor_units', '')::bigint
           OR actual.applied_basis_points IS DISTINCT FROM
                NULLIF(expected->>'applied_basis_points', '')::integer;
        SELECT COALESCE(pg_catalog.sum(pg_catalog.abs(
            (expected->>'opening_value')::bigint - COALESCE(
                CASE expected->>'record_kind'
                    WHEN 'usage_opening' THEN totals.committed_total
                    ELSE totals.gauge_total END, -1
            )
        )), 0) INTO v_usage
        FROM (
            SELECT value || '{"record_kind":"usage_opening"}'::jsonb ||
                pg_catalog.jsonb_build_object('opening_value', value->'count') AS expected
            FROM pg_catalog.jsonb_array_elements(v_run.plan_document->'usage_openings')
            UNION ALL
            SELECT value || '{"record_kind":"gauge_opening"}'::jsonb ||
                pg_catalog.jsonb_build_object('opening_value', value->'value')
            FROM pg_catalog.jsonb_array_elements(v_run.plan_document->'gauge_openings')
        ) AS expected
        LEFT JOIN platform.billing_usage_totals AS totals
          ON totals.meter_period_id = (expected->>'meter_period_id')::uuid;
    END IF; UPDATE platform.billing_v1_adoption_categories AS category SET
        target_count = (SELECT pg_catalog.count(current_sha)
            FROM (SELECT platform.billing_v1_adoption_fingerprint_v1(
                    mapping.record_kind, mapping.target_id) AS current_sha
                  FROM platform.billing_v1_adoption_mappings AS mapping
                  WHERE mapping.operation_id = p_operation_id
                    AND mapping.run_id = p_run_id
                    AND mapping.category = category.category) AS rows),
        baseline_sha256 = (SELECT pg_catalog.sha256(pg_catalog.convert_to(
            COALESCE(pg_catalog.string_agg(mapping.record_kind || ':' ||
                mapping.source_key || ':' || pg_catalog.encode(
                    mapping.record_sha256, 'hex'), '|' ORDER BY
                mapping.record_kind, mapping.source_key), ''), 'UTF8'))
            FROM platform.billing_v1_adoption_mappings AS mapping
            WHERE mapping.operation_id = p_operation_id AND mapping.run_id = p_run_id
              AND mapping.category = category.category),
        target_sha256 = (SELECT pg_catalog.sha256(pg_catalog.convert_to(
            COALESCE(pg_catalog.string_agg(mapping.record_kind || ':' ||
                mapping.source_key || ':' || pg_catalog.encode(
                    platform.billing_v1_adoption_fingerprint_v1(
                        mapping.record_kind, mapping.target_id), 'hex'), '|'
                    ORDER BY mapping.record_kind, mapping.source_key), ''), 'UTF8'))
            FROM platform.billing_v1_adoption_mappings AS mapping
            WHERE mapping.operation_id = p_operation_id AND mapping.run_id = p_run_id
              AND mapping.category = category.category), updated_at = v_now
    WHERE category.operation_id = p_operation_id AND category.run_id = p_run_id;
    UPDATE platform.billing_v1_adoption_runs SET
        state = CASE WHEN v_source_count = v_target_count AND v_baseline = v_target
            AND v_money = 0 AND v_period = 0 AND v_coupon = 0 AND v_usage = 0
            THEN 'verified' ELSE 'blocked' END,
        baseline_sha256 = v_baseline, target_sha256 = v_target,
        money_difference = v_money, period_difference = v_period,
        coupon_difference = v_coupon, usage_difference = v_usage,
        verified_at = v_now, updated_at = v_now
    WHERE operation_id = p_operation_id AND run_id = p_run_id;
    RETURN QUERY SELECT v_source_count, v_target_count,
        pg_catalog.encode(v_baseline, 'hex'), pg_catalog.encode(v_target, 'hex'),
        v_money, v_period, v_coupon, v_usage;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_verify_v1_adoption_v1(uuid, uuid)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_verify_v1_adoption_v1(uuid, uuid)
    TO forgecrud_adopter;
-- +goose StatementEnd
