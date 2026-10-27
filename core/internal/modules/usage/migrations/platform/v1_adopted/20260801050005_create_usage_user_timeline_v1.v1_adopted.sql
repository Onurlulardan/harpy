-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_user_timeline_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text, p_subject text,
    p_page integer, p_limit integer
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_total bigint;
    v_items jsonb;
    v_expected text;
BEGIN
    v_expected := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.user_timeline'
        ELSE 'usage.platform.user_timeline' END;
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_page NOT BETWEEN 1 AND 1000 OR p_limit NOT BETWEEN 1 AND 100
       OR (p_scope = 'tenant' AND p_subject !~
           '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')
       OR (p_scope = 'platform' AND p_subject !~ '^[0-9a-f]{64}$') THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage timeline query is invalid';
    END IF;
    SELECT pg_catalog.count(*) INTO v_total
    FROM platform.usage_request_facts AS facts
    WHERE facts.environment_id = p_environment_id
      AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
      AND ((p_scope = 'tenant' AND facts.actor_id = p_subject::uuid)
        OR (p_scope = 'platform' AND pg_catalog.encode(pg_catalog.sha256(
            pg_catalog.uuid_send(facts.environment_id) || pg_catalog.uuid_send(facts.tenant_id)
            || pg_catalog.uuid_send(facts.actor_id)), 'hex') = p_subject));
    SELECT COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object('id', page.fact_id::text,
            'route_pattern', page.route_pattern, 'method', page.method,
            'status_code', page.status_code,
            'duration_millis', page.duration_micros::numeric / 1000.0,
            'occurred_at', page.occurred_at)
        ORDER BY page.occurred_at DESC, page.fact_id DESC
    ), '[]'::jsonb) INTO v_items
    FROM (
        SELECT facts.fact_id, facts.route_pattern, facts.method, facts.status_code,
            facts.duration_micros, facts.occurred_at
        FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id
          AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
          AND ((p_scope = 'tenant' AND facts.actor_id = p_subject::uuid)
            OR (p_scope = 'platform' AND pg_catalog.encode(pg_catalog.sha256(
                pg_catalog.uuid_send(facts.environment_id) || pg_catalog.uuid_send(facts.tenant_id)
                || pg_catalog.uuid_send(facts.actor_id)), 'hex') = p_subject))
        ORDER BY facts.occurred_at DESC, facts.fact_id DESC
        OFFSET (p_page - 1) * p_limit LIMIT p_limit
    ) AS page;
    RETURN pg_catalog.jsonb_build_object('items', v_items, 'total', v_total,
        'page', p_page, 'limit', p_limit,
        'total_pages', ((v_total + p_limit - 1) / p_limit)::integer);
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_user_timeline_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,text,
integer,integer) FROM PUBLIC;
-- +goose StatementEnd
