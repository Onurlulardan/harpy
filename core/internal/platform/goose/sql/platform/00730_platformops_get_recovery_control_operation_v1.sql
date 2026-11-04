-- +goose Up
CREATE FUNCTION platform.platformops_get_recovery_control_operation_v1(
    in_operation_id text
)
RETURNS TABLE (
    operation_id text,
    environment_id text,
    operation_kind text,
    state text,
    kit_claim_state text,
    artifact_sha256 text,
    expires_at timestamptz,
    failure_code text,
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        NULL::text,
        NULL::timestamptz,
        NULL::text,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.platformops_get_recovery_control_operation_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_recovery_control_operation_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_recovery_control_operation_v1(text);
