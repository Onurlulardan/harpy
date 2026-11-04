-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_grant_acl_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 text,
    arg_5 bytea,
    arg_6 uuid,
    arg_7 text,
    arg_8 uuid,
    arg_9 text,
    arg_10 bigint,
    arg_11 timestamptz
)
RETURNS TABLE (
    dashboard_id uuid,
    principal_kind text,
    principal_id uuid,
    access_level text,
    row_revision bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_grant_acl_v1(name, uuid, uuid, text, bytea, uuid, text, uuid, text, bigint, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_grant_acl_v1(name, uuid, uuid, text, bytea, uuid, text, uuid, text, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_grant_acl_v1(name, uuid, uuid, text, bytea, uuid, text, uuid, text, bigint, timestamptz);
