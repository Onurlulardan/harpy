-- +goose Up
CREATE FUNCTION platform.platformops_get_alert_rule_v1(
    in_environment_id uuid,
    in_rule_id uuid
)
RETURNS TABLE (
    id text,
    environment_id text,
    name text,
    description text,
    enabled boolean,
    service_name text,
    condition text,
    cooldown_minutes bigint,
    channels bytea,
    created_at timestamptz,
    updated_at timestamptz
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
        false,
        NULL::text,
        ''::text,
        0::bigint,
        ''::bytea,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.platformops_get_alert_rule_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_alert_rule_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_get_alert_rule_v1(uuid, uuid);
