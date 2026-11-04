-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_request_recovery_set_v2(
    in_recovery_set_id text,
    in_operation_id text,
    in_environment_id text,
    in_created_at timestamptz
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_request_recovery_set_v2(text, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_request_recovery_set_v2(text, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_request_recovery_set_v2(text, text, text, timestamptz);
