-- +goose Up
CREATE FUNCTION platform.platformops_get_trace_v1(
    in_environment_id uuid,
    in_correlation_id text
)
RETURNS TABLE (
    id text,
    environment_id text,
    correlation_id text,
    span_id text,
    parent_span_id text,
    service_name text,
    operation text,
    span_kind text,
    method text,
    path text,
    status_code bigint,
    started_at timestamptz,
    ended_at timestamptz,
    duration_ms double precision,
    level text,
    message text,
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        NULL::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        0::double precision,
        ''::text,
        NULL::text,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.platformops_get_trace_v1(uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_trace_v1(uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_trace_v1(uuid, text);
