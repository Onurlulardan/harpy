-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_maintenance_unknown_resolve_v1(
    p_scope_schema name,p_kind text,p_tenant_id uuid,p_tenant_schema name,
    p_export_id uuid,p_authority_group_id text,p_authority_group_generation bigint,
    p_cutover_generation bigint,p_job_id uuid,p_attempt_id uuid,
    p_attempt_number integer,p_lease_generation bigint,
    p_original_lease_owner_id uuid,p_effect_id uuid,p_request_sha256 bytea,
    p_unknown_at timestamptz,p_cleanup_generation bigint,
    p_cleanup_lease_owner_id uuid,p_delete_attempt_generation bigint,
    p_resolution_receipt_id uuid,p_effect_receipt_id uuid,
    p_effect_receipt_sha256 bytea,p_resolution_kind text,
    p_resolution_sha256 bytea,p_terminal_state text,p_local_state_sha256 bytea,
    p_local_state_at timestamptz,p_object_key text,p_object_version text,
    p_reconcile_request_id text,p_reconcile_sha256 bytea,
    p_upload_completed_at timestamptz,p_reconcile_verified_at timestamptz,
    p_delete_request_id text,p_delete_response_sha256 bytea,
    p_absence_sha256 bytea,p_delete_verified_at timestamptz,
    p_resolved_at timestamptz
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_resolution analytics_dashboard_export_maintenance_unknown_resolutions%ROWTYPE;
    v_platform_replayed boolean;
    v_platform_effect_receipt_id uuid;
    v_platform_state text;
    v_release_bytes bigint:=0;
    v_active boolean:=false;
    v_mode text;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_authority_group_id!~'^[a-z][a-z0-9_]{0,62}$' OR
       p_authority_group_generation<1 OR p_cutover_generation<1 OR
       p_delete_attempt_generation NOT BETWEEN 0 AND 3 OR
       pg_catalog.uuid_extract_version(p_resolution_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 5 OR
       p_unknown_at IS NULL OR NOT pg_catalog.isfinite(p_unknown_at) OR
       p_resolved_at<=p_unknown_at OR
       analytics_dashboard_export_maintenance_resolution_valid_v1(
          p_kind,p_resolution_kind,p_resolution_sha256,p_terminal_state,
          p_local_state_sha256,p_local_state_at,p_object_key,p_object_version,
          p_reconcile_request_id,p_reconcile_sha256,p_upload_completed_at,
          p_reconcile_verified_at,p_delete_request_id,p_delete_response_sha256,
          p_absence_sha256,p_delete_verified_at,p_resolved_at) IS NOT TRUE THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance resolution scope is invalid';
    END IF;
    SELECT resolved.replayed,resolved.effect_receipt_id,resolved.job_state
    INTO STRICT v_platform_replayed,v_platform_effect_receipt_id,v_platform_state
    FROM platform.platform_async_analytics_dashboard_export_maintenance_unknown_resolve_v1(
        p_kind,p_tenant_id,p_tenant_schema,p_export_id,p_authority_group_id,
        p_authority_group_generation,p_cutover_generation,p_job_id,p_attempt_id,
        p_attempt_number,p_lease_generation,p_original_lease_owner_id,p_effect_id,
        p_request_sha256,p_unknown_at,p_cleanup_generation,p_cleanup_lease_owner_id,
        p_resolution_receipt_id,p_effect_receipt_id,p_effect_receipt_sha256,
        p_terminal_state,p_resolution_sha256,p_resolved_at
    ) AS resolved;
    IF v_platform_effect_receipt_id IS DISTINCT FROM p_effect_receipt_id OR
       v_platform_state IS DISTINCT FROM 'succeeded' THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance platform resolution differs';
    END IF;
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND ((p_kind='analytics.dashboard.export.delete' AND
            item.delete_job_id=p_job_id AND item.expiry_job_id IS DISTINCT FROM p_job_id) OR
           (p_kind='analytics.dashboard.export.expire' AND
            item.expiry_job_id=p_job_id AND item.delete_job_id=p_job_id))
    FOR UPDATE;
    SELECT stored.* INTO v_resolution
    FROM analytics_dashboard_export_maintenance_unknown_resolutions AS stored
    WHERE stored.job_id=p_job_id FOR SHARE;
    IF FOUND THEN
        IF NOT v_platform_replayed OR NOT
           analytics_dashboard_export_maintenance_resolution_replay_matches_v1(
              v_resolution,pg_catalog.jsonb_build_array(
                p_tenant_id,p_tenant_schema,p_kind,p_export_id,p_authority_group_id,
                p_authority_group_generation,p_cutover_generation,p_job_id,p_attempt_id,
                p_attempt_number,p_lease_generation,p_original_lease_owner_id,
                p_effect_id,p_request_sha256,p_unknown_at,p_cleanup_generation,
                p_cleanup_lease_owner_id,p_delete_attempt_generation,
                p_resolution_receipt_id,p_effect_receipt_id,p_effect_receipt_sha256,
                p_resolution_kind,p_resolution_sha256,p_terminal_state,
                p_local_state_sha256,p_local_state_at,p_object_key,p_object_version,
                p_reconcile_request_id,p_reconcile_sha256,p_upload_completed_at,
                p_reconcile_verified_at,p_delete_request_id,p_delete_response_sha256,
                p_absence_sha256,p_delete_verified_at,p_resolved_at)) OR
           v_export.state IS DISTINCT FROM p_terminal_state OR
           v_export.quota_state IS DISTINCT FROM 'released' OR
           v_export.maintenance_unknown_job_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export maintenance resolution replay differs';
        END IF;
        RETURN p_terminal_state;
    END IF;
    IF v_export.export_id IS NULL OR v_platform_replayed OR
       v_export.maintenance_unknown_job_id IS DISTINCT FROM p_job_id OR
       v_export.maintenance_cleanup_generation IS DISTINCT FROM p_cleanup_generation OR
       v_export.maintenance_cleanup_lease_owner IS DISTINCT FROM p_cleanup_lease_owner_id OR
       v_export.maintenance_cleanup_lease_until<=pg_catalog.clock_timestamp() OR
       v_export.maintenance_local_state_sha256 IS DISTINCT FROM p_local_state_sha256 OR
       v_export.maintenance_local_state_at IS DISTINCT FROM p_local_state_at OR
       v_export.delete_attempt_generation IS DISTINCT FROM p_delete_attempt_generation OR
       p_resolved_at<v_export.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance lease differs';
    END IF;
    v_mode:=analytics_dashboard_export_maintenance_resolution_mode_v1(
        v_export,p_kind,p_terminal_state,p_resolution_kind,p_object_key,
        p_object_version,p_delete_request_id,p_delete_response_sha256,
        p_absence_sha256,p_delete_verified_at,p_unknown_at,p_local_state_at,
        p_cleanup_lease_owner_id,p_delete_attempt_generation);
    IF v_mode IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance state differs';
    END IF;
    v_active:=v_mode='active';
    IF v_active THEN
        v_release_bytes:=v_export.ciphertext_bytes;
        UPDATE platform.artifact_quota_counters
        SET used_ciphertext_bytes=used_ciphertext_bytes-v_release_bytes,
            revision=revision+1,updated_at=p_resolved_at
        WHERE tenant_id=p_tenant_id AND used_ciphertext_bytes>=v_release_bytes;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6105',
                MESSAGE='Analytics dashboard export maintenance quota differs';
        END IF;
    END IF;
    INSERT INTO analytics_dashboard_export_maintenance_unknown_resolutions(
        resolution_receipt_id,effect_receipt_id,effect_receipt_sha256,resolution_sha256,
        tenant_id,tenant_schema,kind,export_id,authority_group_id,
        authority_group_generation,cutover_generation,job_id,attempt_id,attempt_number,
        lease_generation,original_lease_owner_id,effect_id,effect_request_sha256,
        original_unknown_at,cleanup_generation,cleanup_lease_owner_id,
        delete_attempt_generation,resolution_kind,terminal_state,local_state_sha256,
        local_state_at,object_key,object_version,reconcile_provider_request_id,
        reconcile_evidence_sha256,upload_completed_at,reconcile_verified_at,
        delete_provider_request_id,delete_response_sha256,absence_sha256,
        delete_verified_at,quota_released_bytes,resolved_at
    ) VALUES (
        p_resolution_receipt_id,p_effect_receipt_id,p_effect_receipt_sha256,
        p_resolution_sha256,p_tenant_id,p_tenant_schema,p_kind,p_export_id,
        p_authority_group_id,p_authority_group_generation,p_cutover_generation,
        p_job_id,p_attempt_id,p_attempt_number,p_lease_generation,
        p_original_lease_owner_id,p_effect_id,p_request_sha256,p_unknown_at,
        p_cleanup_generation,p_cleanup_lease_owner_id,p_delete_attempt_generation,
        p_resolution_kind,p_terminal_state,p_local_state_sha256,p_local_state_at,
        p_object_key,p_object_version,p_reconcile_request_id,p_reconcile_sha256,
        p_upload_completed_at,p_reconcile_verified_at,p_delete_request_id,
        p_delete_response_sha256,p_absence_sha256,p_delete_verified_at,
        v_release_bytes,p_resolved_at
    );
    UPDATE analytics_dashboard_exports
    SET state=CASE WHEN v_active THEN 'deleted' ELSE state END,
        quota_state=CASE WHEN v_active THEN 'released' ELSE quota_state END,
        delete_provider_request_id=CASE WHEN v_active THEN p_delete_request_id
            ELSE delete_provider_request_id END,
        delete_response_sha256=CASE WHEN v_active THEN p_delete_response_sha256
            ELSE delete_response_sha256 END,
        delete_absence_sha256=CASE WHEN v_active THEN p_absence_sha256
            ELSE delete_absence_sha256 END,
        delete_verified_at=CASE WHEN v_active THEN p_delete_verified_at
            ELSE delete_verified_at END,
        deleted_at=CASE WHEN v_active THEN p_resolved_at ELSE deleted_at END,
        delete_lease_owner=NULL,delete_lease_until=NULL,
        maintenance_unknown_job_id=NULL,maintenance_cleanup_generation=NULL,
        maintenance_cleanup_lease_owner=NULL,maintenance_cleanup_lease_until=NULL,
        maintenance_local_state_sha256=NULL,maintenance_local_state_at=NULL,
        row_revision=row_revision+1,updated_at=p_resolved_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN p_terminal_state;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_unknown_resolve_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,bigint,uuid,bigint,uuid,uuid,bytea,text,
    bytea,text,bytea,timestamptz,text,text,text,bytea,timestamptz,
    timestamptz,text,bytea,bytea,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_maintenance_unknown_resolve_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,bigint,uuid,bigint,uuid,uuid,bytea,text,
    bytea,text,bytea,timestamptz,text,text,text,bytea,timestamptz,
    timestamptz,text,bytea,bytea,timestamptz,timestamptz
) TO forgecrud_runtime;
