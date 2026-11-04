-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_import_recovery_snapshot_v1(
    arg_1 text
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::void;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_import_recovery_snapshot_v1(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_import_recovery_snapshot_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_import_recovery_snapshot_v1(text);
