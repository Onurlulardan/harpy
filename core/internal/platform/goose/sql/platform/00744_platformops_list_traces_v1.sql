-- +goose Up
CREATE FUNCTION platform.platformops_list_traces_v1(
    in_environment_id uuid,
    in_service_name text,
    in_operation text,
    in_level text,
    in_status_code integer,
    in_min_status integer,
    in_max_status integer,
    in_minimum_duration numeric,
    in_maximum_duration numeric,
    in_started_after timestamptz,
    in_ended_before timestamptz,
    in_search text,
    in_sort_by text,
    in_sort_direction text,
    in_limit integer,
    in_offset integer
)
RETURNS TABLE (
    correlation_id text,
    first_span timestamptz,
    last_span timestamptz,
    total_duration_millis double precision,
    span_count bigint,
    services text[],
    root_operation text,
    root_status_code bigint,
    has_errors boolean,
    total bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        0::double precision,
        0::bigint,
        '{}'::text[],
        ''::text,
        0::bigint,
        false,
        0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.platformops_list_traces_v1(uuid, text, text, text, integer, integer, integer, numeric, numeric, timestamptz, timestamptz, text, text, text, integer, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_list_traces_v1(uuid, text, text, text, integer, integer, integer, numeric, numeric, timestamptz, timestamptz, text, text, text, integer, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_list_traces_v1(uuid, text, text, text, integer, integer, integer, numeric, numeric, timestamptz, timestamptz, text, text, text, integer, integer);
