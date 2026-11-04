-- +goose Up
CREATE FUNCTION platform.platformops_list_alert_events_v1(
    in_environment_id uuid,
    in_service_name text,
    in_status text,
    in_limit integer,
    in_offset integer
)
RETURNS TABLE (
    id text,
    environment_id text,
    alert_rule_id text,
    service_name text,
    previous_status text,
    current_status text,
    channels_used text[],
    message text,
    resolved_at timestamptz,
    created_at timestamptz,
    total_count bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        ''::text,
        NULL::text,
        ''::text,
        ''::text,
        ''::text,
        '{}'::text[],
        ''::text,
        NULL::timestamptz,
        pg_catalog.clock_timestamp(),
        0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.platformops_list_alert_events_v1(uuid, text, text, integer, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_list_alert_events_v1(uuid, text, text, integer, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_list_alert_events_v1(uuid, text, text, integer, integer);
