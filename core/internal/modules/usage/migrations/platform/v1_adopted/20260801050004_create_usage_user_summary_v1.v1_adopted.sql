-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_user_summary_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text, p_subject text,
    p_start timestamptz, p_end timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_total bigint;
    v_unique bigint;
    v_last timestamptz;
    v_routes jsonb;
    v_daily jsonb;
    v_expected text;
BEGIN
    v_expected := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.user_summary'
        ELSE 'usage.platform.user_summary' END;
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_start IS NULL OR p_end IS NULL OR p_start >= p_end
       OR p_end - p_start > interval '31 days'
       OR (p_scope = 'tenant' AND p_subject !~
           '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
       OR (p_scope = 'platform' AND p_subject !~ '^[0-9a-f]{64}$') THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage subject query is invalid';
    END IF;
    SELECT pg_catalog.count(*),
        pg_catalog.count(DISTINCT (facts.route_pattern, facts.method)),
        pg_catalog.max(facts.occurred_at)
    INTO v_total, v_unique, v_last
    FROM platform.usage_request_facts AS facts
    WHERE facts.environment_id = p_environment_id
      AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
      AND facts.occurred_at >= p_start AND facts.occurred_at < p_end
      AND ((p_scope = 'tenant' AND facts.actor_id = p_subject::uuid)
        OR (p_scope = 'platform' AND pg_catalog.encode(pg_catalog.sha256(
            pg_catalog.uuid_send(facts.environment_id)
            || pg_catalog.uuid_send(facts.tenant_id)
            || pg_catalog.uuid_send(facts.actor_id)), 'hex') = p_subject));
    IF v_total = 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6904', MESSAGE = 'usage subject not found';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object('route_pattern', ranked.route_pattern,
            'method', ranked.method, 'request_count', ranked.request_count,
            'average_millis', ranked.average_millis)
        ORDER BY ranked.request_count DESC, ranked.route_pattern, ranked.method
    ), '[]'::jsonb) INTO v_routes
    FROM (
        SELECT facts.route_pattern, facts.method, pg_catalog.count(*) AS request_count,
            pg_catalog.avg(facts.duration_micros)::numeric / 1000.0 AS average_millis
        FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id
          AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
          AND facts.occurred_at >= p_start AND facts.occurred_at < p_end
          AND ((p_scope = 'tenant' AND facts.actor_id = p_subject::uuid)
            OR (p_scope = 'platform' AND pg_catalog.encode(pg_catalog.sha256(
                pg_catalog.uuid_send(facts.environment_id) || pg_catalog.uuid_send(facts.tenant_id)
                || pg_catalog.uuid_send(facts.actor_id)), 'hex') = p_subject))
        GROUP BY facts.route_pattern, facts.method
        ORDER BY request_count DESC, facts.route_pattern, facts.method LIMIT 10
    ) AS ranked;
    SELECT COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object('date', daily.day, 'request_count', daily.request_count)
        ORDER BY daily.day), '[]'::jsonb) INTO v_daily
    FROM (
        SELECT pg_catalog.to_char(pg_catalog.date_trunc('day', facts.occurred_at, 'UTC'),
            'YYYY-MM-DD') AS day, pg_catalog.count(*) AS request_count
        FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id
          AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
          AND facts.occurred_at >= p_start AND facts.occurred_at < p_end
          AND ((p_scope = 'tenant' AND facts.actor_id = p_subject::uuid)
            OR (p_scope = 'platform' AND pg_catalog.encode(pg_catalog.sha256(
                pg_catalog.uuid_send(facts.environment_id) || pg_catalog.uuid_send(facts.tenant_id)
                || pg_catalog.uuid_send(facts.actor_id)), 'hex') = p_subject))
        GROUP BY 1 ORDER BY 1
    ) AS daily;
    RETURN pg_catalog.jsonb_build_object('subject_id', p_subject, 'last_active', v_last,
        'total_requests', v_total, 'unique_routes', v_unique,
        'top_routes', v_routes, 'daily_breakdown', v_daily);
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_user_summary_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,text,
timestamptz,timestamptz) FROM PUBLIC;
-- +goose StatementEnd
