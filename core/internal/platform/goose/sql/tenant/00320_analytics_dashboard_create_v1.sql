-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_create_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 text,
    arg_5 bytea,
    arg_6 uuid,
    arg_7 uuid,
    arg_8 text,
    arg_9 text,
    arg_10 bytea,
    arg_11 jsonb,
    arg_12 timestamptz
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
    updated_at timestamptz,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
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
        pg_catalog.clock_timestamp(),
        false;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_create_v1(name, uuid, uuid, text, bytea, uuid, uuid, text, text, bytea, jsonb, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_create_v1(name, uuid, uuid, text, bytea, uuid, uuid, text, text, bytea, jsonb, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_create_v1(name, uuid, uuid, text, bytea, uuid, uuid, text, text, bytea, jsonb, timestamptz);
