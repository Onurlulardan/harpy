-- +goose Up
-- BELIRSIZ: offering snapshot rows are not determined by the caller.
CREATE FUNCTION platform.billing_catalog_offering_snapshot_read_v1()
RETURNS TABLE (
    id uuid,
    offering_key text,
    lifecycle text,
    publication_state text,
    is_public boolean,
    available boolean,
    revision bigint,
    version_id uuid,
    version bigint,
    name text,
    description text,
    sort_order integer,
    highlight_label text,
    color text,
    billing_interval text,
    is_base boolean,
    trial_days integer,
    grace_days integer,
    expected_price_count smallint,
    expected_grant_count smallint,
    effective_start timestamptz,
    effective_end timestamptz,
    created_at timestamptz,
    updated_at timestamptz,
    prices jsonb,
    entitlements jsonb
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::uuid,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::boolean,
        NULL::boolean,
        NULL::bigint,
        NULL::uuid,
        NULL::bigint,
        NULL::text,
        NULL::text,
        NULL::integer,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::boolean,
        NULL::integer,
        NULL::integer,
        NULL::smallint,
        NULL::smallint,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::jsonb,
        NULL::jsonb
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.billing_catalog_offering_snapshot_read_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_catalog_offering_snapshot_read_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_catalog_offering_snapshot_read_v1();
