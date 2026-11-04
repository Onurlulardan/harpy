-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_get_alert_event_v1(
    in_environment_id uuid,
    in_event_id uuid
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
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
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
        pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_get_alert_event_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_alert_event_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_alert_event_v1(uuid, uuid);
