-- +goose Up
-- BELIRSIZ: promotion discount rows are not determined by the caller.
CREATE FUNCTION platform.billing_promotion_discount_read_v2()
RETURNS TABLE (
    discount_id uuid,
    legacy_id uuid,
    created_by uuid,
    code text,
    revision bigint,
    version_id uuid,
    version bigint,
    lifecycle text,
    name text,
    description text,
    available boolean,
    discount_kind text,
    basis_points bigint,
    fixed_minor_units bigint,
    currency text,
    currency_registry_version bigint,
    duration text,
    repeating_cycles bigint,
    priority integer,
    all_offerings boolean,
    offering_version_ids uuid[],
    offering_ids uuid[],
    legacy_offering_ids uuid[],
    used_count bigint,
    campaign_ids uuid[],
    legacy_campaign_ids uuid[],
    valid_start timestamptz,
    valid_end timestamptz,
    global_use_limit bigint,
    per_tenant_use_limit bigint,
    created_at timestamptz,
    updated_at timestamptz,
    authority_id uuid
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::text,
        NULL::bigint,
        NULL::uuid,
        NULL::bigint,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::boolean,
        NULL::text,
        NULL::bigint,
        NULL::bigint,
        NULL::text,
        NULL::bigint,
        NULL::text,
        NULL::bigint,
        NULL::integer,
        NULL::boolean,
        NULL::uuid[],
        NULL::uuid[],
        NULL::uuid[],
        NULL::bigint,
        NULL::uuid[],
        NULL::uuid[],
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::bigint,
        NULL::bigint,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::uuid
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.billing_promotion_discount_read_v2() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_promotion_discount_read_v2() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_promotion_discount_read_v2();
