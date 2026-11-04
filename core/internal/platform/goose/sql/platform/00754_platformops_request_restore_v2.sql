-- +goose Up
CREATE FUNCTION platform.platformops_request_restore_v2(
    in_operation_id text,
    in_recovery_set_id text,
    in_environment_id text,
    in_disposition text,
    in_plan_sha256 text,
    in_plan_bytes bytea,
    in_safety_id text,
    in_safety_sha text,
    in_waiver text,
    in_probe text,
    in_destructive_confirmed boolean,
    in_created_at timestamptz,
    in_requested_by_tenant_id text,
    in_requested_by_actor_id text,
    in_requested_by_session_id text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_request_restore_v2(text, text, text, text, text, bytea, text, text, text, text, boolean, timestamptz, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_request_restore_v2(text, text, text, text, text, bytea, text, text, text, text, boolean, timestamptz, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_request_restore_v2(text, text, text, text, text, bytea, text, text, text, text, boolean, timestamptz, text, text, text);
