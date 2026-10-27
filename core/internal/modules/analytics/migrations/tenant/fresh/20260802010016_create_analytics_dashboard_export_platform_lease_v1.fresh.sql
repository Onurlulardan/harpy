-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_platform_checkpoint_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_lease_owner uuid,p_observed_at timestamptz
)
RETURNS TABLE(lease_until timestamptz,cancellation_requested boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_platform_lease timestamptz; v_cancel boolean;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       p_lease_owner IS NULL OR p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export checkpoint scope is invalid';
    END IF;
    SELECT checkpoint.lease_until,checkpoint.cancellation_requested
    INTO v_platform_lease,v_cancel
    FROM platform.platform_async_analytics_dashboard_export_checkpoint_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_lease_owner,p_observed_at
    ) AS checkpoint;
    IF v_platform_lease IS NULL OR NOT pg_catalog.isfinite(v_platform_lease) OR
       v_platform_lease<=p_observed_at OR v_cancel IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export checkpoint result is invalid';
    END IF;
    PERFORM 1 FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id AND item.state='running' AND p_observed_at>item.updated_at
      AND item.quota_state='held'
      AND item.platform_attempt_id=p_attempt_id
      AND item.attempt_generation=p_attempt_generation
      AND item.lease_owner=p_lease_owner FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export checkpoint is stale';
    END IF;
    UPDATE analytics_dashboard_exports
    SET lease_until=v_platform_lease,
        cancellation_requested_at=CASE WHEN v_cancel THEN
            COALESCE(cancellation_requested_at,p_observed_at)
            ELSE cancellation_requested_at END,
        row_revision=row_revision+1,
        updated_at=GREATEST(updated_at,p_observed_at)
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    UPDATE analytics_dashboard_export_platform_attempts
    SET lease_until=v_platform_lease
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation
      AND platform_attempt_id=p_attempt_id AND outcome='running';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export checkpoint evidence changed';
    END IF;
    lease_until:=v_platform_lease;cancellation_requested:=v_cancel;
    RETURN NEXT;
END;
$function$;
CREATE FUNCTION analytics_dashboard_export_platform_effect_start_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,p_object_key text,p_key_id text,
    p_wrapped_dek bytea,p_wrap_nonce bytea,
    p_header_sha256 bytea,p_ciphertext_sha256 bytea,p_plaintext_sha256 bytea,
    p_ciphertext_bytes bigint,p_plaintext_bytes bigint,p_row_count bigint,
    p_content_type text,p_filename text,p_started_at timestamptz
)
RETURNS timestamptz LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid; v_export analytics_dashboard_exports%ROWTYPE;
    v_persisted_at timestamptz; v_exact boolean;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR p_lease_owner IS NULL OR
       pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32 OR
       COALESCE(pg_catalog.octet_length(p_object_key),0) NOT BETWEEN 10 AND 512 OR
       COALESCE(pg_catalog.octet_length(p_key_id),0) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_wrapped_dek) IS DISTINCT FROM 48 OR
       pg_catalog.octet_length(p_wrap_nonce) IS DISTINCT FROM 12 OR
       pg_catalog.octet_length(p_header_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_ciphertext_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_plaintext_sha256) IS DISTINCT FROM 32 OR
       COALESCE(p_ciphertext_bytes,0) NOT BETWEEN 35 AND 2147483648 OR
       COALESCE(p_plaintext_bytes,0) NOT BETWEEN 1 AND 67108864 OR
       COALESCE(p_row_count,-1)<0 OR COALESCE(pg_catalog.octet_length(p_content_type),0) NOT BETWEEN 3 AND 128 OR
       COALESCE(pg_catalog.octet_length(p_filename),0) NOT BETWEEN 5 AND 255 OR
       p_started_at IS NULL OR NOT pg_catalog.isfinite(p_started_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export effect scope is invalid';
    END IF;
    v_persisted_at:=
        platform.platform_async_analytics_dashboard_export_effect_start_v1(
            p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
            p_lease_owner,p_effect_id,p_effect_request_sha256,p_started_at);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id AND item.state='running'
      AND item.quota_state='held' AND item.platform_attempt_id=p_attempt_id
      AND item.attempt_generation=p_attempt_generation
      AND item.lease_owner=p_lease_owner FOR UPDATE;
    IF NOT FOUND OR v_persisted_at IS NULL OR NOT pg_catalog.isfinite(v_persisted_at) OR
       v_export.cancellation_requested_at IS NOT NULL OR v_export.expires_at<=p_started_at OR
       (v_export.effect_started_at IS NULL AND v_persisted_at<=v_export.updated_at) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export effect lease is stale';
    END IF;
    IF v_export.effect_started_at IS NOT NULL THEN
        SELECT receipt.platform_attempt_id IS NOT DISTINCT FROM p_attempt_id AND
               receipt.platform_effect_id IS NOT DISTINCT FROM p_effect_id AND
               receipt.platform_effect_request_sha256 IS NOT DISTINCT FROM p_effect_request_sha256 AND
               receipt.object_key IS NOT DISTINCT FROM p_object_key AND
               receipt.key_id IS NOT DISTINCT FROM p_key_id AND
               receipt.wrapped_dek IS NOT DISTINCT FROM p_wrapped_dek AND
               receipt.wrap_nonce IS NOT DISTINCT FROM p_wrap_nonce AND
               receipt.header_sha256 IS NOT DISTINCT FROM p_header_sha256 AND
               receipt.ciphertext_sha256 IS NOT DISTINCT FROM p_ciphertext_sha256 AND
               receipt.plaintext_sha256 IS NOT DISTINCT FROM p_plaintext_sha256 AND
               receipt.ciphertext_bytes IS NOT DISTINCT FROM p_ciphertext_bytes AND
               receipt.plaintext_bytes IS NOT DISTINCT FROM p_plaintext_bytes AND
               receipt.row_count IS NOT DISTINCT FROM p_row_count AND
               v_export.content_type IS NOT DISTINCT FROM p_content_type AND
               v_export.filename IS NOT DISTINCT FROM p_filename AND
               receipt.started_at IS NOT DISTINCT FROM v_persisted_at INTO v_exact
        FROM analytics_dashboard_export_attempt_receipts AS receipt
        WHERE receipt.export_id=p_export_id
          AND receipt.attempt_generation=p_attempt_generation FOR UPDATE;
        IF v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
           v_export.platform_effect_request_sha256 IS DISTINCT FROM p_effect_request_sha256 OR
           v_export.effect_started_at IS DISTINCT FROM v_persisted_at OR
           NOT COALESCE(v_exact,false) THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export effect replay differs';
        END IF;
        RETURN v_persisted_at;
    END IF;
    UPDATE analytics_dashboard_exports
    SET effect_started_at=v_persisted_at,platform_effect_id=p_effect_id,
        platform_effect_request_sha256=p_effect_request_sha256,
        object_key=p_object_key,key_id=p_key_id,wrapped_dek=p_wrapped_dek,
        wrap_nonce=p_wrap_nonce,header_sha256=p_header_sha256,
        ciphertext_sha256=p_ciphertext_sha256,
        plaintext_sha256=p_plaintext_sha256,
        ciphertext_bytes=p_ciphertext_bytes,plaintext_bytes=p_plaintext_bytes,
        row_count=p_row_count,content_type=p_content_type,filename=p_filename,
        row_revision=row_revision+1,updated_at=v_persisted_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    UPDATE analytics_dashboard_export_platform_attempts
    SET platform_effect_id=p_effect_id,
        platform_effect_request_sha256=p_effect_request_sha256,
        effect_started_at=v_persisted_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation
      AND platform_attempt_id=p_attempt_id AND outcome='running';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export attempt evidence changed';
    END IF;
    INSERT INTO analytics_dashboard_export_attempt_receipts(
        export_id,attempt_generation,object_key,key_id,wrapped_dek,wrap_nonce,
        header_sha256,ciphertext_sha256,plaintext_sha256,ciphertext_bytes,
        plaintext_bytes,row_count,outcome,started_at,platform_attempt_id,
        platform_effect_id,platform_effect_request_sha256
    ) VALUES (
        p_export_id,p_attempt_generation,p_object_key,p_key_id,p_wrapped_dek,
        p_wrap_nonce,p_header_sha256,p_ciphertext_sha256,p_plaintext_sha256,
        p_ciphertext_bytes,p_plaintext_bytes,p_row_count,'started',
        v_persisted_at,p_attempt_id,p_effect_id,p_effect_request_sha256
    );
    RETURN v_persisted_at;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_checkpoint_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_effect_start_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,text,text,bytea,bytea,
    bytea,bytea,bytea,bigint,bigint,bigint,text,text,timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_checkpoint_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,timestamptz) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_effect_start_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,text,text,bytea,bytea,
    bytea,bytea,bytea,bigint,bigint,bigint,text,text,timestamptz) TO forgecrud_runtime;
