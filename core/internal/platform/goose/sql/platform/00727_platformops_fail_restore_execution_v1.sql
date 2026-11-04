-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_fail_restore_execution_v1(
    in_operation_id text,
    in_attempt_id text,
    in_lease_token text,
    in_code text,
    in_fenced boolean
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

REVOKE ALL ON FUNCTION platform.platformops_fail_restore_execution_v1(text, text, text, text, boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_fail_restore_execution_v1(text, text, text, text, boolean) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_fail_restore_execution_v1(text, text, text, text, boolean);
