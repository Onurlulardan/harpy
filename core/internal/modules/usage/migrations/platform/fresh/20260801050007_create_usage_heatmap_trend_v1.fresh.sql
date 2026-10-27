-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_heatmap_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text,
    p_start timestamptz, p_end timestamptz
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_result jsonb; v_expected text;
BEGIN
    v_expected := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.heatmap'
        ELSE 'usage.platform.heatmap' END;
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_start IS NULL OR p_end IS NULL OR p_start >= p_end
       OR p_end - p_start > interval '31 days' THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage heatmap query is invalid';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object('day', cells.day, 'hour', cells.hour, 'count', cells.count)
        ORDER BY cells.day, cells.hour), '[]'::jsonb) INTO v_result
    FROM (
        SELECT pg_catalog.date_part('dow', facts.occurred_at AT TIME ZONE 'UTC')::integer AS day,
            pg_catalog.date_part('hour', facts.occurred_at AT TIME ZONE 'UTC')::integer AS hour,
            pg_catalog.count(*) AS count
        FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id
          AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
          AND facts.occurred_at >= p_start AND facts.occurred_at < p_end
        GROUP BY 1, 2 ORDER BY 1, 2
    ) AS cells;
    RETURN v_result;
END;
$$;

CREATE FUNCTION platform.usage_trend_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text,
    p_start timestamptz, p_end timestamptz, p_granularity text
) RETURNS jsonb LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE v_result jsonb; v_expected text;
BEGIN
    v_expected := CASE WHEN p_scope = 'tenant' THEN 'usage.tenant.trend'
        ELSE 'usage.platform.trend' END;
    PERFORM platform.usage_assert_read_access_v1(
        p_actor_id, p_subject_tenant_id, p_target_tenant_id, p_environment_id,
        p_token_version, p_authorization_version, p_authority_clock, p_scope,
        p_mode, p_purpose, p_control_primary, p_permission, p_operation, v_expected);
    IF p_start IS NULL OR p_end IS NULL OR p_start >= p_end
       OR p_end - p_start > interval '31 days'
       OR p_granularity NOT IN ('hour','day','week') THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage trend query is invalid';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object('bucket', points.bucket, 'requests', points.requests,
            'unique_users', points.unique_users) ORDER BY points.bucket), '[]'::jsonb)
    INTO v_result
    FROM (
        SELECT pg_catalog.date_trunc(p_granularity, facts.occurred_at, 'UTC') AS bucket,
            pg_catalog.count(*) AS requests,
            pg_catalog.count(DISTINCT facts.actor_id) AS unique_users
        FROM platform.usage_request_facts AS facts
        WHERE facts.environment_id = p_environment_id
          AND (p_scope = 'platform' OR facts.tenant_id = p_target_tenant_id)
          AND facts.occurred_at >= p_start AND facts.occurred_at < p_end
        GROUP BY 1 ORDER BY 1
    ) AS points;
    RETURN v_result;
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_heatmap_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_trend_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,
timestamptz,timestamptz,text) FROM PUBLIC;
-- +goose StatementEnd
