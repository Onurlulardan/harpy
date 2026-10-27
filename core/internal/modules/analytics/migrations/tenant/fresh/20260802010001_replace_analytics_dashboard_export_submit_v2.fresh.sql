-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_dashboard_export_submit_v1(
    p_scope_schema name, p_operation_id uuid, p_actor_id uuid,
    p_idempotency_key text, p_request_sha256 bytea,
    p_export_id uuid, p_job_id uuid, p_dashboard_id uuid,
    p_format text, p_parameters bytea,
    p_token_version bigint, p_authorization_version bigint,
    p_authorization_clock bigint, p_maximum_rows bigint,
    p_maximum_plaintext_bytes bigint, p_occurred_at timestamptz,
    p_expires_at timestamptz
)
RETURNS TABLE (
    export_id uuid, dashboard_id uuid, dashboard_version_id uuid, job_id uuid,
    effective_state text, stored_state text, format text, filename text,
    content_type text, failure_code text, row_count bigint,
    plaintext_bytes bigint, cancellation_requested boolean,
    created_at timestamptz, updated_at timestamptz, expires_at timestamptz,
    replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing analytics_dashboard_exports%ROWTYPE;
    v_version analytics_dashboard_versions%ROWTYPE;
    v_tenant_id uuid;
    v_widgets jsonb;
    v_widget_count integer;
    v_pin_sha256 bytea;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 16 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_format NOT IN ('csv','xlsx','pdf') OR
       pg_catalog.octet_length(p_parameters) NOT BETWEEN 2 AND 65536 OR
       NOT pg_catalog.pg_input_is_valid(
           pg_catalog.convert_from(p_parameters,'UTF8'),'jsonb') OR
       p_token_version<1 OR p_authorization_version<1 OR
       p_authorization_clock<1 OR p_maximum_rows NOT BETWEEN 1 AND 100000 OR
       p_maximum_plaintext_bytes NOT BETWEEN 1 AND 67108864 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       p_expires_at IS NULL OR NOT pg_catalog.isfinite(p_expires_at) OR
       p_expires_at<=p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export submit input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=pg_catalog.current_schema();
    INSERT INTO platform.artifact_quota_counters(tenant_id)
    VALUES (v_tenant_id) ON CONFLICT (tenant_id) DO NOTHING;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:export:'||p_actor_id::text||':'||
        p_idempotency_key,6100));
    SELECT item.* INTO v_existing
    FROM analytics_dashboard_exports AS item
    WHERE item.actor_id=p_actor_id
      AND item.idempotency_key=p_idempotency_key
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 OR
           v_existing.tenant_id IS DISTINCT FROM v_tenant_id THEN
            RAISE EXCEPTION USING ERRCODE='P6103',
                MESSAGE='analytics dashboard export replay differs';
        END IF;
        RETURN QUERY
        SELECT result.*,true
        FROM analytics_dashboard_export_snapshot_v1(
            p_scope_schema,p_actor_id,v_existing.dashboard_id,
            v_existing.export_id,p_occurred_at) AS result;
        RETURN;
    END IF;
    SELECT version_row.* INTO v_version
    FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.dashboard_id=root.id
    WHERE root.id=p_dashboard_id
      AND root.lifecycle_state='active'
      AND version_row.version_state='published'
    ORDER BY version_row.version_no DESC
    LIMIT 1
    FOR SHARE OF root,version_row;
    IF NOT FOUND OR NOT analytics_dashboard_export_access_v1(
        p_actor_id,p_dashboard_id,v_version.id) THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics published dashboard was not found';
    END IF;
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
               'widget_id',widget.widget_id,
               'position',widget.position,
               'query_version_id',widget.query_version_id,
               'parameters',pg_catalog.encode(widget.parameter_document,'hex'),
               'output',pg_catalog.encode(widget.output_document,'hex')
           ) ORDER BY widget.position),'[]'::jsonb),
           pg_catalog.count(*)::integer
    INTO v_widgets,v_widget_count
    FROM analytics_dashboard_widgets AS widget
    WHERE widget.dashboard_version_id=v_version.id;
    IF v_widget_count NOT BETWEEN 1 AND 30 THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard export widgets are unavailable';
    END IF;
    v_pin_sha256:=pg_catalog.sha256(pg_catalog.convert_to(
        pg_catalog.jsonb_build_object(
            'dashboard_id',p_dashboard_id,
            'dashboard_version_id',v_version.id,
            'dashboard_manifest_sha256',
                pg_catalog.encode(v_version.manifest_sha256,'hex'),
            'format',p_format,
            'parameters',pg_catalog.encode(p_parameters,'hex'),
            'widgets',v_widgets,
            'maximum_rows',p_maximum_rows,
            'maximum_plaintext_bytes',p_maximum_plaintext_bytes
        )::text,'UTF8'));
    INSERT INTO analytics_dashboard_exports (
        export_id,tenant_id,dashboard_id,dashboard_version_id,job_id,actor_id,
        format,state,row_revision,idempotency_key,request_sha256,
        token_version,authorization_version,authorization_clock,
        dashboard_manifest_sha256,pin_sha256,parameters_document,
        parameters_sha256,pinned_widgets,widget_count,maximum_rows,
        maximum_plaintext_bytes,quota_state,created_at,updated_at,expires_at
    ) VALUES (
        p_export_id,v_tenant_id,p_dashboard_id,v_version.id,p_job_id,p_actor_id,
        p_format,'queued',1,p_idempotency_key,p_request_sha256,
        p_token_version,p_authorization_version,p_authorization_clock,
        v_version.manifest_sha256,v_pin_sha256,p_parameters,
        pg_catalog.sha256(p_parameters),v_widgets,v_widget_count,
        p_maximum_rows,p_maximum_plaintext_bytes,'none',
        p_occurred_at,p_occurred_at,p_expires_at
    );
    RETURN QUERY
    SELECT result.*,false
    FROM analytics_dashboard_export_snapshot_v1(
        p_scope_schema,p_actor_id,p_dashboard_id,p_export_id,p_occurred_at) AS result;
END;
$function$;
-- +goose StatementEnd

