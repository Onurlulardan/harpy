-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_record_trace_v1(
    in_new_string uuid,
    in_environment_id uuid,
    in_correlation_id text,
    in_span_id text,
    in_arg_5 text,
    in_service_name text,
    in_operation text,
    in_span_kind text,
    in_method text,
    in_path text,
    in_status_code integer,
    in_started_at timestamptz,
    in_ended_at timestamptz,
    in_duration_millis numeric,
    in_level text,
    in_arg_16 text,
    in_ended_at_17 timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_record_trace_v1(uuid, uuid, text, text, text, text, text, text, text, text, integer, timestamptz, timestamptz, numeric, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_record_trace_v1(uuid, uuid, text, text, text, text, text, text, text, text, integer, timestamptz, timestamptz, numeric, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_record_trace_v1(uuid, uuid, text, text, text, text, text, text, text, text, integer, timestamptz, timestamptz, numeric, text, text, timestamptz);
