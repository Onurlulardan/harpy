-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_start_restore_verify_v1(
    in_operation_id text,
    in_attempt_id text,
    in_lease_token text,
    in_archive_readback_sha256 text,
    in_new_database_system_id text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_start_restore_verify_v1(text, text, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_start_restore_verify_v1(text, text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_start_restore_verify_v1(text, text, text, text, text);
