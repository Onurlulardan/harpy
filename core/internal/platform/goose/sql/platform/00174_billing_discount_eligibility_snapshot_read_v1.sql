-- +goose Up
-- BELIRSIZ: eligibility snapshot is not determined by the caller.
CREATE FUNCTION platform.billing_discount_eligibility_snapshot_read_v1(
    p_tenant_authority_id uuid,
    p_tenant_id uuid,
    p_code text,
    p_offering_version_id uuid,
    p_at timestamptz,
    p_global_authority_id uuid
)
RETURNS TABLE (
    tenant_id uuid,
    candidate_count bigint,
    discount_id uuid,
    revision bigint,
    discount_version_id uuid,
    version bigint,
    subscription_id uuid,
    offering_version_id uuid,
    price_version_id uuid,
    subscription_currency text,
    normalized_code text,
    discount_kind text,
    fixed_minor_units bigint,
    currency text,
    currency_registry_version bigint,
    basis_points integer,
    duration text,
    repeating_cycles bigint,
    priority integer,
    global_use_limit bigint,
    per_tenant_use_limit bigint,
    all_offerings boolean,
    expected_offering_count smallint,
    valid_start timestamptz,
    valid_end timestamptz,
    available boolean,
    global_used bigint,
    tenant_used bigint,
    offering_allowed boolean,
    already_redeemed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::uuid,
        NULL::bigint,
        NULL::uuid,
        NULL::bigint,
        NULL::uuid,
        NULL::bigint,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::bigint,
        NULL::text,
        NULL::bigint,
        NULL::integer,
        NULL::text,
        NULL::bigint,
        NULL::integer,
        NULL::bigint,
        NULL::bigint,
        NULL::boolean,
        NULL::smallint,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::boolean,
        NULL::bigint,
        NULL::bigint,
        NULL::boolean,
        NULL::boolean
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.billing_discount_eligibility_snapshot_read_v1(
    uuid, uuid, text, uuid, timestamptz, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_discount_eligibility_snapshot_read_v1(
    uuid, uuid, text, uuid, timestamptz, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_discount_eligibility_snapshot_read_v1(
    uuid, uuid, text, uuid, timestamptz, uuid
);
