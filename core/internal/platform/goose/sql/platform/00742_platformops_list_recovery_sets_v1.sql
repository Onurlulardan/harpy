-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_list_recovery_sets_v1(
    in_environment_id text
)
RETURNS TABLE (
    recovery_set_id text,
    operation_id text,
    environment_id text,
    sequence bigint,
    state text,
    manifest_sha text,
    legal_hold boolean,
    database_system_id text,
    protected_at timestamptz
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
        0::bigint,
        ''::text,
        NULL::text,
        false,
        NULL::text,
        NULL::timestamptz;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_list_recovery_sets_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_list_recovery_sets_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_list_recovery_sets_v1(text);
