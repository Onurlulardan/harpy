-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_maintenance_unknown_claim_v1(
    p_scope_schema name,p_kind text,p_tenant_id uuid,p_tenant_schema name,
    p_export_id uuid,p_authority_group_id text,p_authority_group_generation bigint,
    p_cutover_generation bigint,p_job_id uuid,p_attempt_id uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_unknown_at timestamptz,
    p_cleanup_lease_owner_id uuid,p_observed_at timestamptz
)
RETURNS TABLE(
    claim_state text,cleanup_generation bigint,lease_until timestamptz,
    cleanup_mode text,terminal_state text,format text,object_key text,
    object_version text,key_id text,wrapped_dek bytea,wrap_nonce bytea,
    header_sha256 bytea,ciphertext_sha256 bytea,plaintext_sha256 bytea,
    ciphertext_bytes bigint,plaintext_bytes bigint,row_count bigint,
    content_type text,filename text,render_generation bigint,
    delete_attempt_generation bigint,
    local_state_sha256 bytea,local_state_at timestamptz,
    delete_provider_request_id text,delete_response_sha256 bytea,
    delete_absence_sha256 bytea,delete_verified_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_claim_state text;
    v_cleanup_generation bigint;
    v_lease_until timestamptz;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_mode text;
    v_terminal text;
    v_local_at timestamptz;
    v_local_sha bytea;
    v_delete_generation bigint;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_kind NOT IN ('analytics.dashboard.export.delete',
                      'analytics.dashboard.export.expire') OR
       p_authority_group_id!~'^[a-z][a-z0-9_]{0,62}$' OR
       p_authority_group_generation<1 OR p_cutover_generation<1 OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_number NOT BETWEEN 1 AND
          (CASE p_kind WHEN 'analytics.dashboard.export.delete' THEN 3 ELSE 8 END) OR
       p_lease_generation IS DISTINCT FROM p_attempt_number OR
       pg_catalog.uuid_extract_version(p_original_lease_owner_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 5 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_unknown_at IS NULL OR NOT pg_catalog.isfinite(p_unknown_at) OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance claim scope is invalid';
    END IF;
    SELECT claimed.claim_state,claimed.cleanup_generation,claimed.lease_until
    INTO STRICT v_claim_state,v_cleanup_generation,v_lease_until
    FROM platform.platform_async_analytics_dashboard_export_maintenance_unknown_claim_v1(
        p_kind,p_tenant_id,p_tenant_schema,p_export_id,p_authority_group_id,
        p_authority_group_generation,p_cutover_generation,p_job_id,p_attempt_id,
        p_attempt_number,p_lease_generation,p_original_lease_owner_id,p_effect_id,
        p_request_sha256,p_unknown_at,p_cleanup_lease_owner_id,p_observed_at
    ) AS claimed;
    IF v_claim_state IN ('busy','resolved') THEN
        RETURN QUERY SELECT v_claim_state,NULL::bigint,NULL::timestamptz,
            NULL::text,NULL::text,NULL::text,NULL::text,NULL::text,NULL::text,
            NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,NULL::bytea,
            NULL::bigint,NULL::bigint,NULL::bigint,NULL::text,NULL::text,
            NULL::bigint,NULL::bigint,NULL::bytea,NULL::timestamptz,NULL::text,NULL::bytea,
            NULL::bytea,NULL::timestamptz;
        RETURN;
    END IF;
    IF v_claim_state IS DISTINCT FROM 'claimed' OR v_cleanup_generation<1 OR
       v_lease_until<=p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance claim differs';
    END IF;
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND ((p_kind='analytics.dashboard.export.delete' AND
            item.delete_job_id=p_job_id AND item.expiry_job_id IS DISTINCT FROM p_job_id) OR
           (p_kind='analytics.dashboard.export.expire' AND
            item.expiry_job_id=p_job_id AND item.delete_job_id=p_job_id AND
            item.delete_operation_id=p_job_id))
    FOR UPDATE;
    IF v_export.export_id IS NULL OR EXISTS (
        SELECT 1 FROM analytics_dashboard_export_maintenance_unknown_resolutions
        WHERE job_id=p_job_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance local identity differs';
    END IF;
    IF v_export.state='deleted' AND v_export.quota_state='released' THEN
        v_mode:='completed';v_terminal:='deleted';
    ELSIF p_kind='analytics.dashboard.export.expire' AND
          v_export.state IN ('expired','cancelled','failed') AND
          v_export.quota_state='released' AND v_export.object_key IS NULL AND
          v_export.effect_started_at IS NULL THEN
        v_mode:='terminal';v_terminal:=v_export.state;
    ELSIF v_export.state IN ('ready','deleting') AND
          v_export.quota_state='used' AND v_export.object_key IS NOT NULL AND
          v_export.object_version IS NOT NULL AND v_export.key_id IS NOT NULL AND
          v_export.wrapped_dek IS NOT NULL AND v_export.wrap_nonce IS NOT NULL AND
          v_export.header_sha256 IS NOT NULL AND v_export.ciphertext_sha256 IS NOT NULL AND
          v_export.plaintext_sha256 IS NOT NULL AND v_export.ciphertext_bytes IS NOT NULL AND
          v_export.plaintext_bytes IS NOT NULL AND v_export.row_count IS NOT NULL THEN
        v_mode:='artifact';v_terminal:='deleted';
    ELSE
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance local state differs';
    END IF;
    v_delete_generation:=CASE WHEN v_mode='artifact' THEN
        LEAST(v_export.delete_attempt_generation+1,3)
        ELSE v_export.delete_attempt_generation END;
    v_local_at:=GREATEST(p_observed_at,v_export.updated_at+interval '1 microsecond');
    v_local_sha:=pg_catalog.sha256(pg_catalog.convert_to(pg_catalog.jsonb_build_array(
        p_kind,p_export_id,p_job_id,v_cleanup_generation,p_cleanup_lease_owner_id,
        v_mode,v_terminal,v_export.state,v_export.quota_state,v_export.row_revision+1,
        v_local_at,v_export.object_key,v_export.object_version,
        v_delete_generation,p_cleanup_lease_owner_id,v_lease_until,
        v_export.delete_provider_request_id,
        pg_catalog.encode(COALESCE(v_export.delete_response_sha256,''::bytea),'hex'),
        pg_catalog.encode(COALESCE(v_export.delete_absence_sha256,''::bytea),'hex'),
        v_export.delete_verified_at)::text,'UTF8'));
    UPDATE analytics_dashboard_exports
    SET state=CASE WHEN v_mode='artifact' THEN 'deleting' ELSE state END,
        delete_attempt_generation=CASE WHEN v_mode='artifact' THEN
            v_delete_generation ELSE analytics_dashboard_exports.delete_attempt_generation END,
        delete_lease_owner=CASE WHEN v_mode='artifact'
            THEN p_cleanup_lease_owner_id ELSE delete_lease_owner END,
        delete_lease_until=CASE WHEN v_mode='artifact'
            THEN v_lease_until ELSE delete_lease_until END,
        maintenance_unknown_job_id=p_job_id,
        maintenance_cleanup_generation=v_cleanup_generation,
        maintenance_cleanup_lease_owner=p_cleanup_lease_owner_id,
        maintenance_cleanup_lease_until=v_lease_until,
        maintenance_local_state_sha256=v_local_sha,
        maintenance_local_state_at=v_local_at,
        row_revision=row_revision+1,updated_at=v_local_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    SELECT item.* INTO STRICT v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id;
    RETURN QUERY SELECT v_claim_state,v_cleanup_generation,v_lease_until,
        v_mode,v_terminal,v_export.format,v_export.object_key,v_export.object_version,
        v_export.key_id,v_export.wrapped_dek,v_export.wrap_nonce,
        v_export.header_sha256,v_export.ciphertext_sha256,v_export.plaintext_sha256,
        v_export.ciphertext_bytes,v_export.plaintext_bytes,v_export.row_count,
        v_export.content_type,v_export.filename,v_export.attempt_generation,
        v_export.delete_attempt_generation,
        v_local_sha,v_local_at,v_export.delete_provider_request_id,
        v_export.delete_response_sha256,v_export.delete_absence_sha256,
        v_export.delete_verified_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_unknown_claim_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_maintenance_unknown_claim_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,uuid,timestamptz
) TO forgecrud_runtime;
