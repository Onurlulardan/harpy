-- +goose Up
CREATE FUNCTION platform.platformops_test_alert_rule_v1(
    in_new_string uuid,
    in_id uuid,
    in_environment_id uuid,
    in_actor_id uuid,
    in_message text,
    in_utc timestamptz
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
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.platformops_test_alert_rule_v1(uuid, uuid, uuid, uuid, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_test_alert_rule_v1(uuid, uuid, uuid, uuid, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_test_alert_rule_v1(uuid, uuid, uuid, uuid, text, timestamptz);
