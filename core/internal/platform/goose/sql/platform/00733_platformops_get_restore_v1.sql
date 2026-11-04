-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_get_restore_v1(
    in_operation_id text
)
RETURNS TABLE (
    operation_id text,
    recovery_set_id text,
    environment_id text,
    target_disposition text,
    state text,
    plan_sha256 text,
    receipt_sha256 text,
    new_database_system_id text,
    created_at timestamptz,
    updated_at timestamptz,
    completed_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        NULL::text,
        NULL::text,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        NULL::timestamptz;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_get_restore_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_restore_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_restore_v1(text);
