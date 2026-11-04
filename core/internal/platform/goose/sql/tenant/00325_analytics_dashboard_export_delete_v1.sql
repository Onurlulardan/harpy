-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_delete_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 text,
    arg_6 bytea,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 timestamptz
)
RETURNS TABLE (
    export_id uuid,
    dashboard_id uuid,
    dashboard_version_id uuid,
    job_id uuid,
    effective_state text,
    stored_state text,
    format text,
    filename text,
    content_type text,
    failure_code text,
    row_count bigint,
    plaintext_bytes bigint,
    cancellation_requested boolean,
    created_at timestamptz,
    updated_at timestamptz,
    expires_at timestamptz,
    replay boolean,
    schedule boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        0::bigint,
        false,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        false,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_v1(name, uuid, uuid, uuid, text, bytea, uuid, uuid, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_v1(name, uuid, uuid, uuid, text, bytea, uuid, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_delete_v1(name, uuid, uuid, uuid, text, bytea, uuid, uuid, timestamptz);
