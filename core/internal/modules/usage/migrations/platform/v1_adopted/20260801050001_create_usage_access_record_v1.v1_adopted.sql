-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.usage_assert_read_access_v1(
    p_actor_id uuid, p_subject_tenant_id uuid, p_target_tenant_id uuid,
    p_environment_id uuid, p_token_version bigint, p_authorization_version bigint,
    p_authority_clock bigint, p_scope text, p_mode text, p_purpose text,
    p_control_primary boolean, p_permission text, p_operation text,
    p_expected_operation text
) RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_tenant boolean;
    v_platform boolean;
BEGIN
    IF p_actor_id IS NULL OR p_subject_tenant_id IS NULL OR p_environment_id IS NULL
       OR p_token_version < 1 OR p_authorization_version < 1 OR p_authority_clock < 1
       OR p_operation IS DISTINCT FROM p_expected_operation
       OR p_operation !~ '^[a-z][a-z0-9_.]{0,127}$' THEN
        RAISE EXCEPTION USING ERRCODE = 'P6903', MESSAGE = 'usage read access denied';
    END IF;
    v_tenant := p_scope = 'tenant' AND p_target_tenant_id = p_subject_tenant_id
        AND p_permission = 'analytics.usage.read' AND NOT p_control_primary
        AND ((p_mode = 'onpremise' AND p_purpose = 'onpremise_primary')
            OR (p_mode = 'saas' AND p_purpose = 'saas_customer'));
    v_platform := p_scope = 'platform' AND p_target_tenant_id IS NULL
        AND p_permission = 'platform.analytics.usage.read' AND p_control_primary
        AND p_mode = 'saas' AND p_purpose = 'saas_control';
    IF NOT v_tenant AND NOT v_platform THEN
        RAISE EXCEPTION USING ERRCODE = 'P6903', MESSAGE = 'usage read scope denied';
    END IF;
END;
$$;

CREATE FUNCTION platform.usage_record_fact_v1(
    p_fact_id uuid, p_environment_id uuid, p_tenant_id uuid, p_actor_id uuid,
    p_source_identity bytea, p_fact_digest bytea, p_method text, p_route_pattern text,
    p_status_code integer, p_duration_micros bigint, p_occurred_at timestamptz
) RETURNS TABLE(fact_id uuid, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $$
DECLARE
    v_fact_id uuid;
    v_digest bytea;
BEGIN
    IF p_fact_id IS NULL OR pg_catalog.uuid_extract_version(p_fact_id) <> 7
       OR p_environment_id IS NULL OR p_tenant_id IS NULL OR p_actor_id IS NULL
       OR pg_catalog.octet_length(p_source_identity) <> 32
       OR pg_catalog.octet_length(p_fact_digest) <> 32
       OR p_source_identity = pg_catalog.decode(pg_catalog.repeat('00', 32), 'hex')
       OR p_fact_digest = pg_catalog.decode(pg_catalog.repeat('00', 32), 'hex')
       OR p_method NOT IN ('GET','HEAD','POST','PUT','PATCH','DELETE','OPTIONS')
       OR pg_catalog.octet_length(p_route_pattern) NOT BETWEEN 1 AND 256
       OR pg_catalog.left(p_route_pattern, 1) <> '/'
       OR pg_catalog.strpos(p_route_pattern, '?') > 0
       OR pg_catalog.strpos(p_route_pattern, '#') > 0
       OR p_route_pattern ~ E'[\r\n]' OR p_status_code NOT BETWEEN 100 AND 599
       OR p_duration_micros NOT BETWEEN 0 AND 86400000000 OR p_occurred_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6901', MESSAGE = 'usage fact is invalid';
    END IF;
    INSERT INTO platform.usage_request_facts AS target (
        fact_id, environment_id, tenant_id, actor_id, source_identity, fact_digest,
        method, route_pattern, status_code, duration_micros, occurred_at, created_at
    ) VALUES (
        p_fact_id, p_environment_id, p_tenant_id, p_actor_id, p_source_identity,
        p_fact_digest, p_method, p_route_pattern, p_status_code, p_duration_micros,
        p_occurred_at, pg_catalog.clock_timestamp()
    ) ON CONFLICT (environment_id, source_identity) DO NOTHING
    RETURNING target.fact_id INTO v_fact_id;
    IF v_fact_id IS NOT NULL THEN
        fact_id := v_fact_id;
        replayed := false;
        RETURN NEXT;
        RETURN;
    END IF;
    SELECT stored.fact_id, stored.fact_digest INTO v_fact_id, v_digest
    FROM platform.usage_request_facts AS stored
    WHERE stored.environment_id = p_environment_id
      AND stored.source_identity = p_source_identity;
    IF v_fact_id IS NULL OR v_digest IS DISTINCT FROM p_fact_digest THEN
        RAISE EXCEPTION USING ERRCODE = 'P6902', MESSAGE = 'usage source identity conflict';
    END IF;
    fact_id := v_fact_id;
    replayed := true;
    RETURN NEXT;
END;
$$;

REVOKE ALL ON FUNCTION platform.usage_assert_read_access_v1(
uuid,uuid,uuid,uuid,bigint,bigint,bigint,text,text,text,boolean,text,text,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.usage_record_fact_v1(
uuid,uuid,uuid,uuid,bytea,bytea,text,text,integer,bigint,timestamptz) FROM PUBLIC;
-- +goose StatementEnd
