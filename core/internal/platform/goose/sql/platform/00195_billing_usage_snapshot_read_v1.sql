-- +goose Up
-- BELIRSIZ: usage meter rows are not determined by the caller.
CREATE FUNCTION platform.billing_usage_snapshot_read_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_base_offering_version_id uuid,
    p_subscription_id uuid,
    p_subscription_offering_version_id uuid
)
RETURNS TABLE (
    period_count bigint,
    override_count bigint,
    period_id uuid,
    entitlement_key text,
    value_type text,
    meter_kind text,
    reset_policy text,
    enforcement text,
    unlimited_allowed boolean,
    period_start timestamptz,
    period_end timestamptz,
    period_revision bigint,
    committed_total bigint,
    reserved_total bigint,
    gauge_total bigint,
    total_revision bigint,
    updated_at timestamptz,
    limit_amount bigint,
    unlimited boolean,
    limit_present boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::bigint,
        NULL::bigint,
        NULL::uuid,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::boolean,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::bigint,
        NULL::bigint,
        NULL::bigint,
        NULL::bigint,
        NULL::bigint,
        NULL::timestamptz,
        NULL::bigint,
        NULL::boolean,
        NULL::boolean
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.billing_usage_snapshot_read_v1(
    uuid, uuid, uuid, uuid, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_usage_snapshot_read_v1(
    uuid, uuid, uuid, uuid, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_usage_snapshot_read_v1(
    uuid, uuid, uuid, uuid, uuid
);
