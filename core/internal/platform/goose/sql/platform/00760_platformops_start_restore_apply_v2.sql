-- +goose Up
CREATE FUNCTION platform.platformops_start_restore_apply_v2(
    in_operation_id text,
    in_attempt_id text,
    in_lease_token text,
    in_drain_ack_sha256 text,
    in_control_database_locator_sha256 text,
    in_scratch_database_locator_sha256 text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_start_restore_apply_v2(text, text, text, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_start_restore_apply_v2(text, text, text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_start_restore_apply_v2(text, text, text, text, text, text);
