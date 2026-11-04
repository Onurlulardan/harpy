-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_trace_stats_v1(
    in_environment_id uuid
)
RETURNS TABLE (
    total_traces bigint,
    total_spans bigint,
    error_rate double precision,
    average_duration_millis double precision,
    services bytea,
    levels bytea,
    traces_last24_hours bigint,
    traces_last_hour bigint,
    endpoints bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        0::bigint,
        0::double precision,
        0::double precision,
        ''::bytea,
        ''::bytea,
        0::bigint,
        0::bigint,
        ''::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_trace_stats_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_trace_stats_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_trace_stats_v1(uuid);
