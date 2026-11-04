-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_submit_v2(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 text,
    arg_5 bytea,
    arg_6 uuid,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 uuid,
    arg_10 text,
    arg_11 bytea,
    arg_12 bigint,
    arg_13 bigint,
    arg_14 bigint,
    arg_15 bigint,
    arg_16 bigint,
    arg_17 timestamptz,
    arg_18 timestamptz
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
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
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
        false;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_submit_v2(name, uuid, uuid, text, bytea, uuid, uuid, uuid, uuid, text, bytea, bigint, bigint, bigint, bigint, bigint, timestamptz, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_submit_v2(name, uuid, uuid, text, bytea, uuid, uuid, uuid, uuid, text, bytea, bigint, bigint, bigint, bigint, bigint, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_submit_v2(name, uuid, uuid, text, bytea, uuid, uuid, uuid, uuid, text, bytea, bigint, bigint, bigint, bigint, bigint, timestamptz, timestamptz);
