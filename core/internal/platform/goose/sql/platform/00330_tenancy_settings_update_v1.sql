-- +goose Up
CREATE FUNCTION platform.tenancy_settings_update_v1(
    in_tenant_id uuid,
    in_name text,
    in_expected_revision bigint,
    in_idempotency_key text,
    in_request_sha256 bytea
)
RETURNS TABLE (
    result_revision bigint,
    replay boolean,
    outcome text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        0::bigint,
        false,
        ''::text
    ;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_settings_update_v1(uuid, text, bigint, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_settings_update_v1(uuid, text, bigint, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_settings_update_v1(uuid, text, bigint, text, bytea);
