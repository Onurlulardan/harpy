-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_features_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text,
    p_start timestamptz, p_end timestamptz, p_limit integer
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_result jsonb;
    v_expected text;
BEGIN
    v_expected := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.features'
        ELSE 'usage.platform.features' END;
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_start IS NULL OR p_end IS NULL OR p_start >= p_end
       OR p_end - p_start > interval '31 days' OR p_limit NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage features query is invalid';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object('route_pattern', ranked.route_pattern,
            'method', ranked.method, 'request_count', ranked.request_count,
            'unique_users', ranked.unique_users,
            'average_millis', ranked.average_millis,
            'error_count', ranked.error_count,
            'error_rate', ranked.error_count::numeric * 100.0 / ranked.request_count)
        ORDER BY ranked.request_count DESC, ranked.route_pattern, ranked.method
    ), '[]'::jsonb) INTO v_result
    FROM (
        SELECT facts.route_pattern, facts.method,
            pg_catalog.count(*) AS request_count,
            pg_catalog.count(DISTINCT facts.actor_id) AS unique_users,
            pg_catalog.avg(facts.duration_micros)::numeric / 1000.0 AS average_millis,
            pg_catalog.count(*) FILTER (WHERE facts.status_code >= 400) AS error_count
        FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id
          AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
          AND facts.occurred_at >= p_start AND facts.occurred_at < p_end
        GROUP BY facts.route_pattern, facts.method
        ORDER BY request_count DESC, facts.route_pattern, facts.method
        LIMIT p_limit
    ) AS ranked;
    RETURN v_result;
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_features_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,integer) FROM PUBLIC;
-- +goose StatementEnd
