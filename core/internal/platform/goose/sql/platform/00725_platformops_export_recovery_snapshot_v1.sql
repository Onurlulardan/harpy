-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_export_recovery_snapshot_v1(
    in_arg_1 text
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT ''::text;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_export_recovery_snapshot_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_export_recovery_snapshot_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_export_recovery_snapshot_v1(text);
