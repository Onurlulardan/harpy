-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_expiry_control_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 timestamptz
)
RETURNS TABLE (
    state text,
    cleanup_required boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::text,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_expiry_control_v1(name, uuid, uuid, uuid, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_expiry_control_v1(name, uuid, uuid, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_expiry_control_v1(name, uuid, uuid, uuid, timestamptz);
