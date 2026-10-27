-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_top_users_v1(
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
    v_expected := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.top_users'
        ELSE 'usage.platform.top_users' END;
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_start IS NULL OR p_end IS NULL OR p_start >= p_end
       OR p_end - p_start > interval '31 days' OR p_limit NOT BETWEEN 1 AND 50 THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage top users query is invalid';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object(
            'subject_id', ranked.subject_id,
            'request_count', ranked.request_count,
            'unique_routes', ranked.unique_routes,
            'last_active', ranked.last_active
        ) ORDER BY ranked.request_count DESC, ranked.last_active DESC, ranked.subject_id
    ), '[]'::jsonb) INTO v_result
    FROM (
        SELECT CASE WHEN p_scope = 'tenant' THEN facts.actor_id::text ELSE
            pg_catalog.encode(pg_catalog.sha256(
                pg_catalog.uuid_send(facts.environment_id)
                || pg_catalog.uuid_send(facts.tenant_id)
                || pg_catalog.uuid_send(facts.actor_id)), 'hex') END AS subject_id,
            pg_catalog.count(*) AS request_count,
            pg_catalog.count(DISTINCT (facts.route_pattern, facts.method)) AS unique_routes,
            pg_catalog.max(facts.occurred_at) AS last_active
        FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id
          AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
          AND facts.occurred_at >= p_start AND facts.occurred_at < p_end
        GROUP BY 1
        ORDER BY request_count DESC, last_active DESC, subject_id
        LIMIT p_limit
    ) AS ranked;
    RETURN v_result;
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_top_users_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,integer) FROM PUBLIC;
-- +goose StatementEnd
