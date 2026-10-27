-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_list_alert_rules_v1(p_environment_id uuid)
RETURNS SETOF platform.platformops_alert_rules
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT rule.* FROM platform.platformops_alert_rules AS rule
    WHERE rule.environment_id = p_environment_id
    ORDER BY rule.name, rule.id;
END;
$function$;

CREATE FUNCTION platform.platformops_create_alert_rule_v1(
    p_id uuid, p_environment_id uuid, p_name text, p_description text,
    p_enabled boolean, p_service_name text, p_condition text,
    p_cooldown_minutes integer, p_channels jsonb, p_now timestamptz
) RETURNS SETOF platform.platformops_alert_rules
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    INSERT INTO platform.platformops_alert_rules(
        id, environment_id, name, description, enabled, service_name,
        condition, cooldown_minutes, channels, created_at, updated_at
    ) VALUES (
        p_id, p_environment_id, btrim(p_name), p_description, p_enabled,
        nullif(btrim(p_service_name), ''), p_condition, p_cooldown_minutes,
        p_channels, p_now, p_now
    );
    RETURN QUERY SELECT * FROM platform.platformops_alert_rules WHERE id = p_id;
END;
$function$;

CREATE FUNCTION platform.platformops_update_alert_rule_v1(
    p_id uuid, p_environment_id uuid, p_name text, p_description text,
    p_enabled boolean, p_service_name text, p_condition text,
    p_cooldown_minutes integer, p_channels jsonb, p_now timestamptz
) RETURNS SETOF platform.platformops_alert_rules
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_alert_rules SET
        name = btrim(p_name), description = p_description, enabled = p_enabled,
        service_name = nullif(btrim(p_service_name), ''), condition = p_condition,
        cooldown_minutes = p_cooldown_minutes, channels = p_channels,
        revision = revision + 1, updated_at = p_now
    WHERE id = p_id AND environment_id = p_environment_id;
    IF NOT FOUND THEN RAISE EXCEPTION 'alert rule was not found' USING ERRCODE = 'P0002'; END IF;
    RETURN QUERY SELECT * FROM platform.platformops_alert_rules WHERE id = p_id;
END;
$function$;

CREATE FUNCTION platform.platformops_delete_alert_rule_v1(
    p_id uuid, p_environment_id uuid
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    DELETE FROM platform.platformops_alert_rules
    WHERE id = p_id AND environment_id = p_environment_id;
    IF NOT FOUND THEN RAISE EXCEPTION 'alert rule was not found' USING ERRCODE = 'P0002'; END IF;
END;
$function$;

CREATE FUNCTION platform.platformops_test_alert_rule_v1(
    p_event_id uuid, p_rule_id uuid, p_environment_id uuid,
    p_actor_id uuid, p_message text, p_now timestamptz
) RETURNS SETOF platform.platformops_alert_events
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE selected platform.platformops_alert_rules%ROWTYPE;
BEGIN
    SELECT * INTO selected FROM platform.platformops_alert_rules
    WHERE id = p_rule_id AND environment_id = p_environment_id;
    IF NOT FOUND THEN RAISE EXCEPTION 'alert rule was not found' USING ERRCODE = 'P0002'; END IF;
    INSERT INTO platform.platformops_alert_events(
        id, environment_id, alert_rule_id, service_name, previous_status,
        current_status, channels_used, message, actor_id, created_at
    ) VALUES (
        p_event_id, p_environment_id, p_rule_id,
        coalesce(selected.service_name, 'core-api'), 'healthy', 'unhealthy',
        '{}'::text[], p_message, p_actor_id, p_now
    );
    RETURN QUERY SELECT * FROM platform.platformops_alert_events WHERE id = p_event_id;
END;
$function$;

CREATE FUNCTION platform.platformops_list_alert_events_v1(
    p_environment_id uuid, p_service_name text, p_status text,
    p_limit integer, p_offset integer
) RETURNS TABLE(
    id uuid, environment_id uuid, alert_rule_id uuid, service_name text,
    previous_status text, current_status text, channels_used text[], message text,
    actor_id uuid, resolved_at timestamptz, created_at timestamptz, total_count bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_limit NOT BETWEEN 1 AND 100 OR p_offset < 0 THEN
        RAISE EXCEPTION 'alert event query is invalid' USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT event.id, event.environment_id, event.alert_rule_id,
        event.service_name, event.previous_status, event.current_status,
        event.channels_used, event.message, event.actor_id, event.resolved_at,
        event.created_at, count(*) OVER ()
    FROM platform.platformops_alert_events AS event
    WHERE event.environment_id = p_environment_id
      AND (p_service_name = '' OR event.service_name = p_service_name)
      AND (p_status = '' OR event.current_status = p_status)
    ORDER BY event.created_at DESC, event.id
    LIMIT p_limit OFFSET p_offset;
END;
$function$;
-- +goose StatementEnd
