-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_installation_open_usage_v2(
    in_operation_id uuid,
    in_tenant_id uuid,
    in_manifest_sha256 text
)
RETURNS TABLE (
    meter_count integer,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::integer,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_installation_open_usage_v2(uuid, uuid, text);
