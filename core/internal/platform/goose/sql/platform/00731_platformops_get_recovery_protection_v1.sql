-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_get_recovery_protection_v1(
    in_environment_id text
)
RETURNS TABLE (
    environment_id text,
    state text,
    generation bigint,
    kit_acknowledged boolean,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        ''::text,
        0::bigint,
        false,
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_get_recovery_protection_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_recovery_protection_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_recovery_protection_v1(text);
