-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_trace_stats_v1(p_environment_id uuid)
RETURNS TABLE(
    total_traces bigint, total_spans bigint, error_rate numeric,
    average_duration_ms numeric, spans_by_service jsonb, spans_by_level jsonb,
    traces_last_24h bigint, traces_last_hour bigint, slowest_endpoints jsonb
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    WITH source AS (
        SELECT span.* FROM platform.platformops_trace_spans AS span
        WHERE span.environment_id = p_environment_id
          AND span.started_at >= pg_catalog.statement_timestamp() - interval '31 days'
    ), totals AS (
        SELECT count(DISTINCT correlation_id)::bigint AS traces,
            count(*)::bigint AS spans,
            coalesce(avg(duration_ms), 0)::numeric AS average_duration,
            coalesce(count(*) FILTER (WHERE status_code >= 500 OR level IN ('error','fatal'))::numeric
                / nullif(count(*), 0), 0)::numeric AS failures,
            count(DISTINCT correlation_id) FILTER (
                WHERE started_at >= pg_catalog.statement_timestamp() - interval '24 hours')::bigint AS last_day,
            count(DISTINCT correlation_id) FILTER (
                WHERE started_at >= pg_catalog.statement_timestamp() - interval '1 hour')::bigint AS last_hour
        FROM source
    ), services AS (
        SELECT service_name, count(*)::bigint AS amount FROM source GROUP BY service_name
    ), levels AS (
        SELECT level, count(*)::bigint AS amount FROM source GROUP BY level
    ), endpoints AS (
        SELECT path, method, avg(duration_ms)::numeric AS average_duration,
            count(*)::bigint AS calls,
            count(*) FILTER (WHERE status_code >= 500 OR level IN ('error','fatal'))::bigint AS errors
        FROM source GROUP BY path, method ORDER BY average_duration DESC, path, method LIMIT 10
    )
    SELECT totals.traces, totals.spans, totals.failures, totals.average_duration,
        coalesce((SELECT jsonb_object_agg(service_name, amount ORDER BY service_name) FROM services), '{}'::jsonb),
        coalesce((SELECT jsonb_object_agg(level, amount ORDER BY level) FROM levels), '{}'::jsonb),
        totals.last_day, totals.last_hour,
        coalesce((SELECT jsonb_agg(jsonb_build_object(
            'path', path, 'method', method, 'avg_duration_ms', average_duration,
            'call_count', calls, 'error_count', errors
        ) ORDER BY average_duration DESC, path, method) FROM endpoints), '[]'::jsonb)
    FROM totals;
END;
$function$;

CREATE FUNCTION platform.platformops_prune_observability_v1(
    p_environment_id uuid, p_before timestamptz, p_limit integer
) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE deleted_count bigint;
BEGIN
    IF p_limit NOT BETWEEN 1 AND 10000 OR
       p_before > pg_catalog.clock_timestamp() - interval '30 days' THEN
        RAISE EXCEPTION 'observability prune boundary is invalid' USING ERRCODE = '22023';
    END IF;
    WITH candidates AS (
        SELECT id FROM platform.platformops_trace_spans
        WHERE environment_id = p_environment_id AND started_at < p_before
        ORDER BY started_at, id LIMIT p_limit FOR UPDATE SKIP LOCKED
    ), removed AS (
        DELETE FROM platform.platformops_trace_spans AS span
        USING candidates WHERE span.id = candidates.id RETURNING span.id
    ) SELECT count(*) INTO deleted_count FROM removed;
    RETURN deleted_count;
END;
$function$;
-- +goose StatementEnd
