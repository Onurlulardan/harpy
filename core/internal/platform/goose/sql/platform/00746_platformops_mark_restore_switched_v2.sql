-- +goose Up
CREATE FUNCTION platform.platformops_mark_restore_switched_v2(
    in_operation_id text,
    in_attempt_id text,
    in_lease_token text,
    in_object_merkle_sha256 text,
    in_locator_remap_sha256 text,
    in_safety_evidence_sha256 text,
    in_switch_ack_sha256 text,
    in_previous_database_system_id text,
    in_new_database_system_id text,
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

REVOKE ALL ON FUNCTION platform.platformops_mark_restore_switched_v2(text, text, text, text, text, text, text, text, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_mark_restore_switched_v2(text, text, text, text, text, text, text, text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_mark_restore_switched_v2(text, text, text, text, text, text, text, text, text, text, text);
