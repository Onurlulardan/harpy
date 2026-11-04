-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_create_alert_rule_v1(
    in_rule_id uuid,
    in_environment_id uuid,
    in_name text,
    in_description text,
    in_enabled boolean,
    in_service_name text,
    in_condition text,
    in_cooldown_minutes integer,
    in_channels jsonb,
    in_observed_at timestamptz
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
AS $fn$
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
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_create_alert_rule_v1(uuid, uuid, text, text, boolean, text, text, integer, jsonb, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_create_alert_rule_v1(uuid, uuid, text, text, boolean, text, text, integer, jsonb, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_create_alert_rule_v1(uuid, uuid, text, text, boolean, text, text, integer, jsonb, timestamptz);
