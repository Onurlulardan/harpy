SELECT id, environment_id, correlation_id, span_id, parent_span_id,
       service_name, operation, span_kind, method, path, status_code,
       started_at, ended_at, duration_ms, level, message, created_at
FROM platform.platformops_get_trace_span_v1($1::uuid,$2::text)
