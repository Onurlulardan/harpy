-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_overview_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text,
    p_start timestamptz, p_end timestamptz, p_now timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_active jsonb;
    v_features jsonb;
    v_users jsonb;
    v_trend jsonb;
    v_recent jsonb;
    v_expected text;
    v_prefix text;
BEGIN
    v_prefix := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.' ELSE 'usage.platform.' END;
    v_expected := v_prefix || 'overview';
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_start IS NULL OR p_end IS NULL OR p_now IS NULL OR p_start >= p_end
       OR p_end - p_start > interval '31 days' THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage overview query is invalid';
    END IF;
    v_active := platform.usage_active_users_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission,
        v_prefix || 'active_users', p_now);
    v_features := platform.usage_features_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission,
        v_prefix || 'features', p_start, p_end, 5);
    v_users := platform.usage_top_users_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission,
        v_prefix || 'top_users', p_start, p_end, 5);
    v_trend := platform.usage_trend_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission,
        v_prefix || 'trend', p_start, p_end, 'hour');
    SELECT COALESCE(pg_catalog.jsonb_agg(last_points.value
        ORDER BY last_points.ordinality), '[]'::jsonb) INTO v_recent
    FROM (
        SELECT points.value, points.ordinality
        FROM pg_catalog.jsonb_array_elements(v_trend) WITH ORDINALITY AS points(value, ordinality)
        ORDER BY points.ordinality DESC LIMIT 24
    ) AS last_points;
    RETURN pg_catalog.jsonb_build_object('active_users', v_active,
        'top_features', v_features, 'top_users', v_users, 'recent_trend', v_recent);
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_overview_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,timestamptz) FROM PUBLIC;
-- +goose StatementEnd
