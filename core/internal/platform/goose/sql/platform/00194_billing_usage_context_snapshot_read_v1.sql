-- +goose Up
-- BELIRSIZ: usage context counts and offering links are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_usage_context_snapshot_read_v1(
    p_authority_id uuid,
    p_tenant_id uuid
)
RETURNS TABLE (
    base_count bigint,
    base_offering_version_id text,
    subscription_count bigint,
    subscription_id text,
    subscription_offering_version_id text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        NULL::text,
        0::bigint,
        NULL::text,
        NULL::text;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_usage_context_snapshot_read_v1(
    uuid, uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_usage_context_snapshot_read_v1(
    uuid, uuid
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_usage_context_snapshot_read_v1(
    uuid, uuid
);
