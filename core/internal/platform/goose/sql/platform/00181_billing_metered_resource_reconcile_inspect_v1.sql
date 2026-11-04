-- +goose Up
-- BELIRSIZ: stored reconciliation and projection are not determined by the caller.
CREATE FUNCTION platform.billing_metered_resource_reconcile_inspect_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource text,
    p_idempotency_key text,
    p_owner_fact_id uuid
)
RETURNS TABLE (
    receipt_found boolean,
    operation_id uuid,
    request_sha256 bytea,
    evidence_sha256 bytea,
    resource_exists boolean,
    active boolean,
    included_in_opening boolean,
    changed boolean,
    gauge_delta bigint,
    current_value bigint,
    limit_value bigint,
    unlimited boolean,
    recorded_at timestamptz,
    projection_found boolean,
    projection_state text,
    projection_revision bigint,
    projection_updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        false,
        NULL::uuid,
        NULL::bytea,
        NULL::bytea,
        false,
        false,
        false,
        false,
        0::bigint,
        0::bigint,
        0::bigint,
        false,
        NULL::timestamptz,
        false,
        ''::text,
        0::bigint,
        NULL::timestamptz;
$$;

REVOKE ALL ON FUNCTION platform.billing_metered_resource_reconcile_inspect_v1(
    uuid, uuid, uuid, text, text, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_reconcile_inspect_v1(
    uuid, uuid, uuid, text, text, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_metered_resource_reconcile_inspect_v1(
    uuid, uuid, uuid, text, text, uuid
);
