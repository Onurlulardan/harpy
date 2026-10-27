-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_record_trace_v1(
    p_id uuid, p_environment_id uuid, p_correlation_id text,
    p_span_id text, p_parent_span_id text, p_service_name text,
    p_operation text, p_span_kind text, p_method text, p_path text,
    p_status_code integer, p_started_at timestamptz, p_ended_at timestamptz,
    p_duration_ms numeric, p_level text, p_message text, p_created_at timestamptz
) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    INSERT INTO platform.platformops_trace_spans(
        id, environment_id, correlation_id, span_id, parent_span_id,
        service_name, operation, span_kind, method, path, status_code,
        started_at, ended_at, duration_ms, level, message, created_at
    ) VALUES (
        p_id, p_environment_id, p_correlation_id, p_span_id, p_parent_span_id,
        p_service_name, p_operation, p_span_kind, p_method, p_path, p_status_code,
        p_started_at, p_ended_at, p_duration_ms, p_level, p_message, p_created_at
    ) ON CONFLICT (environment_id, span_id) DO NOTHING;
END;
$function$;

CREATE FUNCTION platform.platformops_list_traces_v1(
    p_environment_id uuid, p_service_name text, p_operation text, p_level text,
    p_status_code integer, p_min_status integer, p_max_status integer,
    p_min_duration numeric, p_max_duration numeric,
    p_started_after timestamptz, p_ended_before timestamptz, p_search text,
    p_sort_by text, p_sort_direction text, p_limit integer, p_offset integer
) RETURNS TABLE(
    correlation_id text, first_span timestamptz, last_span timestamptz,
    total_duration_ms numeric, span_count bigint, services text[],
    root_operation text, root_status_code integer, has_errors boolean,
    total_count bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF p_limit NOT BETWEEN 1 AND 100 OR p_offset < 0 OR
       p_sort_by NOT IN ('started_at','duration_ms','span_count') OR
       lower(p_sort_direction) NOT IN ('asc','desc') THEN
        RAISE EXCEPTION 'trace query is invalid' USING ERRCODE = '22023';
    END IF;
    RETURN QUERY
    WITH filtered AS (
        SELECT span.* FROM platform.platformops_trace_spans AS span
        WHERE span.environment_id = p_environment_id
          AND (p_service_name = '' OR span.service_name = p_service_name)
          AND (p_operation = '' OR span.operation = p_operation)
          AND (p_level = '' OR span.level = p_level)
          AND (p_status_code = 0 OR span.status_code = p_status_code)
          AND (p_min_status = 0 OR span.status_code >= p_min_status)
          AND (p_max_status = 0 OR span.status_code <= p_max_status)
          AND (p_min_duration = 0 OR span.duration_ms >= p_min_duration)
          AND (p_max_duration = 0 OR span.duration_ms <= p_max_duration)
          AND (p_started_after IS NULL OR span.started_at >= p_started_after)
          AND (p_ended_before IS NULL OR span.started_at <= p_ended_before)
          AND (p_search = '' OR span.operation ILIKE '%' || p_search || '%'
               OR span.path ILIKE '%' || p_search || '%')
    ), grouped AS (
        SELECT span.correlation_id, min(span.started_at) AS first_span,
            max(span.ended_at) AS last_span,
            extract(epoch FROM (max(span.ended_at) - min(span.started_at))) * 1000 AS duration_ms,
            count(*) AS span_count,
            array_agg(DISTINCT span.service_name ORDER BY span.service_name) AS services,
            (array_agg(span.operation ORDER BY span.started_at, span.span_id))[1] AS root_operation,
            (array_agg(span.status_code ORDER BY span.started_at, span.span_id))[1] AS root_status,
            bool_or(span.status_code >= 500 OR span.level IN ('error','fatal')) AS has_errors
        FROM filtered AS span GROUP BY span.correlation_id
    )
    SELECT grouped.correlation_id, grouped.first_span, grouped.last_span,
        grouped.duration_ms, grouped.span_count, grouped.services,
        grouped.root_operation, grouped.root_status, grouped.has_errors,
        count(*) OVER ()
    FROM grouped
    ORDER BY
      CASE WHEN p_sort_by = 'started_at' AND lower(p_sort_direction) = 'asc' THEN grouped.first_span END ASC,
      CASE WHEN p_sort_by = 'started_at' AND lower(p_sort_direction) = 'desc' THEN grouped.first_span END DESC,
      CASE WHEN p_sort_by = 'duration_ms' AND lower(p_sort_direction) = 'asc' THEN grouped.duration_ms END ASC,
      CASE WHEN p_sort_by = 'duration_ms' AND lower(p_sort_direction) = 'desc' THEN grouped.duration_ms END DESC,
      CASE WHEN p_sort_by = 'span_count' AND lower(p_sort_direction) = 'asc' THEN grouped.span_count END ASC,
      CASE WHEN p_sort_by = 'span_count' AND lower(p_sort_direction) = 'desc' THEN grouped.span_count END DESC,
      grouped.correlation_id
    LIMIT p_limit OFFSET p_offset;
END;
$function$;

CREATE FUNCTION platform.platformops_get_trace_v1(
    p_environment_id uuid, p_correlation_id text
) RETURNS SETOF platform.platformops_trace_spans
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT span.* FROM platform.platformops_trace_spans AS span
    WHERE span.environment_id = p_environment_id
      AND span.correlation_id = p_correlation_id
    ORDER BY span.started_at, span.span_id
    LIMIT 500;
END;
$function$;
-- +goose StatementEnd
