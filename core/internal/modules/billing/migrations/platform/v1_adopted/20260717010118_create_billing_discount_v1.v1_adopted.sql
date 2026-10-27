-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_discount_v1(
    p_authority_id uuid, p_operation_id uuid, p_discount_id uuid,
    p_version_id uuid, p_actor_id uuid, p_idempotency_key text,
    p_request_sha256 bytea, p_code text, p_discount_kind text,
    p_fixed_minor_units bigint, p_currency text,
    p_currency_registry_version bigint, p_basis_points integer,
    p_duration text, p_repeating_cycles bigint, p_priority integer,
    p_global_use_limit bigint, p_per_tenant_use_limit bigint,
    p_all_offerings boolean, p_expected_offering_count smallint,
    p_valid_start timestamptz, p_valid_end timestamptz,
    p_occurred_at timestamptz
)
RETURNS TABLE (
    result_discount_id uuid, result_version_id uuid,
    result_revision bigint, result_version bigint,
    result_lifecycle text, replay boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    stored_request_sha256 bytea;
    stored_aggregate_id uuid;
    stored_version_id uuid;
    stored_revision bigint;
    stored_version bigint;
    stored_lifecycle text;
BEGIN
    PERFORM platform.billing_assert_global_authority_v1(p_authority_id);
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        p_authority_id::text || '|' || p_actor_id::text || '|' || p_idempotency_key, 0
    ));
    SELECT request_sha256, aggregate_id,
           billing_catalog_operations.result_version_id,
           billing_catalog_operations.result_revision,
           billing_catalog_operations.result_version,
           billing_catalog_operations.result_lifecycle
      INTO stored_request_sha256, stored_aggregate_id, stored_version_id,
           stored_revision, stored_version, stored_lifecycle
    FROM platform.billing_catalog_operations
    WHERE authority_id = p_authority_id AND actor_id = p_actor_id
      AND idempotency_key = p_idempotency_key
    FOR SHARE;
    IF FOUND THEN
        IF stored_request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'billing catalog idempotency key was reused';
        END IF;
        RETURN QUERY SELECT stored_aggregate_id, stored_version_id,
            stored_revision, stored_version, stored_lifecycle, true;
        RETURN;
    END IF;
    INSERT INTO platform.billing_discounts (
        id, authority_id, normalized_code, lifecycle,
        created_at, revoked_at, revision
    ) VALUES (
        p_discount_id, p_authority_id, p_code,
        'active', p_occurred_at, NULL, 1
    );
    INSERT INTO platform.billing_discount_versions (
        id, authority_id, discount_id, version, discount_kind,
        fixed_minor_units, currency, currency_registry_version,
        currency_eligible, basis_points, duration, repeating_cycles,
        priority, global_use_limit, per_tenant_use_limit,
        all_offerings, expected_offering_count, publication_operation_id,
        valid_start, valid_end, published_at
    ) VALUES (
        p_version_id, p_authority_id, p_discount_id, 1, p_discount_kind,
        p_fixed_minor_units, p_currency::character(3), p_currency_registry_version,
        CASE WHEN p_discount_kind = 'fixed' THEN true ELSE NULL END,
        p_basis_points, p_duration, p_repeating_cycles, p_priority,
        p_global_use_limit, p_per_tenant_use_limit,
        p_all_offerings, p_expected_offering_count, p_operation_id,
        p_valid_start, p_valid_end, p_occurred_at
    );
    INSERT INTO platform.billing_catalog_operations (
        operation_id, authority_id, actor_id, idempotency_key, request_sha256,
        aggregate_kind, aggregate_id, operation_kind, result_version_id,
        result_revision, result_version, result_lifecycle, reason, occurred_at
    ) VALUES (
        p_operation_id, p_authority_id, p_actor_id, p_idempotency_key,
        p_request_sha256, 'discount', p_discount_id, 'create',
        p_version_id, 1, 1, 'active', NULL, p_occurred_at
    );
    RETURN QUERY SELECT p_discount_id, p_version_id, 1::bigint,
        1::bigint, 'active'::text, false;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_create_discount_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,text,bigint,text,bigint,
    integer,text,bigint,integer,bigint,bigint,boolean,smallint,
    timestamptz,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_discount_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,text,bigint,text,bigint,
    integer,text,bigint,integer,bigint,bigint,boolean,smallint,
    timestamptz,timestamptz,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
