-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_get_alert_rule_v1(
    p_environment_id uuid, p_rule_id uuid
) RETURNS SETOF platform.platformops_alert_rules
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY SELECT rule.*
    FROM platform.platformops_alert_rules AS rule
    WHERE rule.environment_id = p_environment_id
      AND rule.id = p_rule_id;
END;
$function$;

CREATE FUNCTION platform.platformops_get_alert_event_v1(
    p_environment_id uuid, p_event_id uuid
) RETURNS SETOF platform.platformops_alert_events
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY SELECT event.*
    FROM platform.platformops_alert_events AS event
    WHERE event.environment_id = p_environment_id
      AND event.id = p_event_id;
END;
$function$;

CREATE FUNCTION platform.platformops_get_trace_span_v1(
    p_environment_id uuid, p_span_id text
) RETURNS SETOF platform.platformops_trace_spans
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY SELECT span.*
    FROM platform.platformops_trace_spans AS span
    WHERE span.environment_id = p_environment_id
      AND span.span_id = p_span_id;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platformops_get_alert_rule_v1(uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_alert_event_v1(uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_get_trace_span_v1(uuid,text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_get_alert_rule_v1(uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_alert_event_v1(uuid,uuid) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platformops_get_trace_span_v1(uuid,text) TO harpy_runtime;
-- +goose StatementEnd
