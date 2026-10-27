-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_active_users_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text,
    p_now timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_result jsonb;
    v_expected text;
BEGIN
    v_expected := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.active_users'
        ELSE 'usage.platform.active_users' END;
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage time is invalid';
    END IF;
    SELECT pg_catalog.jsonb_build_object(
        'dau', pg_catalog.count(DISTINCT facts.actor_id) FILTER (
            WHERE facts.occurred_at >= pg_catalog.date_trunc('day', p_now, 'UTC')),
        'wau', pg_catalog.count(DISTINCT facts.actor_id) FILTER (
            WHERE facts.occurred_at >= p_now - interval '7 days'),
        'mau', pg_catalog.count(DISTINCT facts.actor_id) FILTER (
            WHERE facts.occurred_at >= p_now - interval '1 month'),
        'concurrent_users', pg_catalog.count(DISTINCT facts.actor_id) FILTER (
            WHERE facts.occurred_at >= p_now - interval '5 minutes'),
        'total_requests_today', pg_catalog.count(*) FILTER (
            WHERE facts.occurred_at >= pg_catalog.date_trunc('day', p_now, 'UTC'))
    ) INTO v_result
    FROM platform.usage_request_facts AS facts
    WHERE facts.environment_id = p_environment_id
      AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
      AND facts.occurred_at <= p_now;
    RETURN v_result;
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_active_users_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,timestamptz)
FROM PUBLIC;
-- +goose StatementEnd
