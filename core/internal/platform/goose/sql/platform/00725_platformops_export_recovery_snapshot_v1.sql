-- +goose Up
CREATE FUNCTION platform.platformops_export_recovery_snapshot_v1(
    in_arg_1 text
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::text;
$$;

REVOKE ALL ON FUNCTION platform.platformops_export_recovery_snapshot_v1(text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_export_recovery_snapshot_v1(text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_export_recovery_snapshot_v1(text);
