-- +goose Up
-- BELIRSIZ: stored mutation and projection are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_metered_resource_inspect_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource text,
    p_mutation text,
    p_idempotency_key text,
    p_owner_fact_id uuid
)
RETURNS TABLE (
    receipt_found boolean,
    operation_id uuid,
    owner_fact_id uuid,
    request_sha256 bytea,
    feature_enabled boolean,
    allowed boolean,
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
AS $fn$
    SELECT
        false,
        NULL::uuid,
        NULL::uuid,
        NULL::bytea,
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
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_metered_resource_inspect_v1(
    uuid, uuid, uuid, text, text, text,
    uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_metered_resource_inspect_v1(
    uuid, uuid, uuid, text, text, text,
    uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_metered_resource_inspect_v1(
    uuid, uuid, uuid, text, text, text,
    uuid
);
