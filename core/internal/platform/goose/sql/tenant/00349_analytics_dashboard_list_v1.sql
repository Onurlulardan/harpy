-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_list_v1(
    arg_1 name,
    arg_2 integer,
    arg_3 uuid
)
RETURNS TABLE (
    dashboard_id uuid,
    version_id uuid,
    version_no bigint,
    root_revision bigint,
    version_revision bigint,
    lifecycle_state text,
    version_state text,
    "name" text,
    visibility text,
    layout_document bytea,
    layout_sha256 bytea,
    manifest_sha256 bytea,
    widgets bytea,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        NULL::bytea,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_list_v1(name, integer, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_list_v1(name, integer, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_list_v1(name, integer, uuid);
