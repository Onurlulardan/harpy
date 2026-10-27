-- +goose Up
ALTER TABLE analytics_dashboard_export_unknown_resolutions
    ADD COLUMN cleanup_generation bigint NOT NULL,
    ADD COLUMN cleanup_lease_owner uuid NOT NULL,
    ADD CONSTRAINT analytics_dashboard_export_unknown_cleanup_check CHECK (
        cleanup_generation>=1 AND
        pg_catalog.uuid_extract_version(cleanup_lease_owner)=7
    );

-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_unknown_resolve_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_original_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,p_cleanup_generation bigint,
    p_cleanup_lease_owner uuid,p_resolution_receipt_id uuid,
    p_resolution_kind text,p_resolution_sha256 bytea,p_object_key text,
    p_object_version text,p_reconcile_provider_request_id text,
    p_reconcile_evidence_sha256 bytea,p_upload_completed_at timestamptz,
    p_reconcile_verified_at timestamptz,p_delete_provider_request_id text,
    p_delete_response_sha256 bytea,p_absence_sha256 bytea,
    p_delete_verified_at timestamptz,p_resolved_at timestamptz
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_attempt analytics_dashboard_export_platform_attempts%ROWTYPE;
    v_resolution analytics_dashboard_export_unknown_resolutions%ROWTYPE;
    v_terminal_state text;
    v_platform boolean;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_resolution_receipt_id) IS DISTINCT FROM 7 OR
       p_resolution_receipt_id IN (
           p_export_id,p_job_id,p_attempt_id,p_original_lease_owner,p_effect_id,
           p_cleanup_lease_owner) OR
       analytics_dashboard_export_unknown_resolution_evidence_valid_v1(
            p_resolution_kind,p_resolution_sha256,p_object_key,p_object_version,
            p_reconcile_provider_request_id,p_reconcile_evidence_sha256,
            p_upload_completed_at,p_reconcile_verified_at,
            p_delete_provider_request_id,p_delete_response_sha256,
            p_absence_sha256,p_delete_verified_at,
            p_resolved_at) IS NOT TRUE THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export unknown resolution scope is invalid';
    END IF;
    v_platform:=platform.platform_async_analytics_dashboard_export_unknown_resolve_v2(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_original_lease_owner,p_effect_id,p_effect_request_sha256,
        p_cleanup_generation,p_cleanup_lease_owner,p_resolution_receipt_id,
        p_resolution_kind,p_resolution_sha256,p_resolved_at
    );
    IF NOT v_platform THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export platform resolution is invalid';
    END IF;
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id FOR UPDATE;
    SELECT attempt.* INTO v_attempt
    FROM analytics_dashboard_export_platform_attempts AS attempt
    WHERE attempt.export_id=p_export_id
      AND attempt.attempt_generation=p_attempt_generation
      AND attempt.platform_attempt_id=p_attempt_id
      AND attempt.job_id=p_job_id
      AND attempt.lease_owner=p_original_lease_owner FOR UPDATE;
    SELECT resolution.* INTO v_resolution
    FROM analytics_dashboard_export_unknown_resolutions AS resolution
    WHERE resolution.export_id=p_export_id
      AND resolution.attempt_generation=p_attempt_generation FOR SHARE;
    IF FOUND THEN
        IF v_resolution.tenant_id IS DISTINCT FROM p_tenant_id OR
           v_resolution.job_id IS DISTINCT FROM p_job_id OR
           v_resolution.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
           v_resolution.lease_owner IS DISTINCT FROM p_original_lease_owner OR
           v_resolution.platform_effect_id IS DISTINCT FROM p_effect_id OR
           v_resolution.platform_effect_request_sha256 IS DISTINCT FROM p_effect_request_sha256 OR
           v_resolution.cleanup_generation IS DISTINCT FROM p_cleanup_generation OR
           v_resolution.cleanup_lease_owner IS DISTINCT FROM p_cleanup_lease_owner OR
           v_resolution.resolution_receipt_id IS DISTINCT FROM p_resolution_receipt_id OR
           v_resolution.resolution_kind IS DISTINCT FROM p_resolution_kind OR
           v_resolution.resolution_sha256 IS DISTINCT FROM p_resolution_sha256 OR
           v_resolution.object_key IS DISTINCT FROM p_object_key OR
           v_resolution.object_version IS DISTINCT FROM p_object_version OR
           v_resolution.reconcile_provider_request_id IS DISTINCT FROM p_reconcile_provider_request_id OR
           v_resolution.reconcile_evidence_sha256 IS DISTINCT FROM p_reconcile_evidence_sha256 OR
           v_resolution.upload_completed_at IS DISTINCT FROM p_upload_completed_at OR
           v_resolution.reconcile_verified_at IS DISTINCT FROM p_reconcile_verified_at OR
           v_resolution.delete_provider_request_id IS DISTINCT FROM p_delete_provider_request_id OR
           v_resolution.delete_response_sha256 IS DISTINCT FROM p_delete_response_sha256 OR
           v_resolution.absence_sha256 IS DISTINCT FROM p_absence_sha256 OR
           v_resolution.delete_verified_at IS DISTINCT FROM p_delete_verified_at OR
           v_resolution.resolved_at IS DISTINCT FROM p_resolved_at OR
           v_export.state IS DISTINCT FROM v_resolution.terminal_state OR
           v_export.quota_state IS DISTINCT FROM 'released' OR
           v_attempt.resolution_receipt_id IS DISTINCT FROM p_resolution_receipt_id THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export resolution replay differs';
        END IF;
        RETURN v_resolution.terminal_state;
    END IF;
    IF v_export.export_id IS NULL OR v_attempt.export_id IS NULL OR
       v_export.state IS DISTINCT FROM 'running' OR
       v_export.quota_state IS DISTINCT FROM 'held' OR
       v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
       v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
       v_export.platform_effect_request_sha256 IS DISTINCT FROM p_effect_request_sha256 OR
       v_export.lease_owner IS NOT NULL OR
       v_export.failure_code IS DISTINCT FROM 'ARTIFACT_OUTCOME_UNKNOWN' OR
       v_attempt.outcome IS DISTINCT FROM 'unknown' OR
       v_attempt.finished_at IS NULL OR v_attempt.finished_at>=p_resolved_at OR
       p_reconcile_verified_at<v_attempt.finished_at OR
       v_attempt.platform_effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.platform_effect_request_sha256 IS DISTINCT FROM p_effect_request_sha256 OR
       v_attempt.resolution_receipt_id IS NOT NULL OR
       p_object_key IS DISTINCT FROM v_export.object_key OR
       p_object_key IS DISTINCT FROM (
            SELECT receipt.object_key
            FROM analytics_dashboard_export_attempt_receipts AS receipt
            WHERE receipt.export_id=p_export_id
              AND receipt.attempt_generation=p_attempt_generation
              AND receipt.outcome='unknown'
              AND receipt.platform_attempt_id=p_attempt_id
              AND receipt.platform_effect_id=p_effect_id
              AND receipt.platform_effect_request_sha256=p_effect_request_sha256) OR
       (v_export.object_version IS NOT NULL AND
            p_object_version IS DISTINCT FROM v_export.object_version) OR
       p_resolved_at<v_export.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export resolution evidence differs';
    END IF;
    v_terminal_state:=CASE
        WHEN v_export.cancellation_requested_at IS NOT NULL THEN 'cancelled'
        WHEN v_export.expires_at<=p_resolved_at THEN 'expired'
        ELSE 'failed' END;
    INSERT INTO analytics_dashboard_export_unknown_resolutions(
        tenant_id,export_id,attempt_generation,job_id,platform_attempt_id,
        lease_owner,platform_effect_id,platform_effect_request_sha256,
        resolution_receipt_id,resolution_kind,resolution_sha256,
        object_key,object_version,reconcile_provider_request_id,
        reconcile_evidence_sha256,upload_completed_at,reconcile_verified_at,
        delete_provider_request_id,delete_response_sha256,absence_sha256,
        delete_verified_at,
        terminal_state,quota_released_bytes,resolved_at,
        cleanup_generation,cleanup_lease_owner
    ) VALUES (
        p_tenant_id,p_export_id,p_attempt_generation,p_job_id,p_attempt_id,
        p_original_lease_owner,p_effect_id,p_effect_request_sha256,
        p_resolution_receipt_id,p_resolution_kind,p_resolution_sha256,
        p_object_key,p_object_version,p_reconcile_provider_request_id,
        p_reconcile_evidence_sha256,p_upload_completed_at,p_reconcile_verified_at,
        p_delete_provider_request_id,p_delete_response_sha256,p_absence_sha256,
        p_delete_verified_at,
        v_terminal_state,2147483648,p_resolved_at,
        p_cleanup_generation,p_cleanup_lease_owner
    );
    UPDATE analytics_dashboard_export_platform_attempts
    SET resolution_receipt_id=p_resolution_receipt_id,
        resolution_kind=p_resolution_kind,resolution_sha256=p_resolution_sha256,
        resolved_at=p_resolved_at
    WHERE export_id=p_export_id AND attempt_generation=p_attempt_generation;
    UPDATE analytics_dashboard_exports
    SET state=v_terminal_state,quota_state='released',
        failure_code='ARTIFACT_OUTCOME_UNKNOWN_RESOLVED',
        platform_attempt_id=NULL,platform_effect_id=NULL,
        platform_effect_request_sha256=NULL,effect_started_at=NULL,
        object_key=NULL,object_version=NULL,key_id=NULL,wrapped_dek=NULL,
        wrap_nonce=NULL,header_sha256=NULL,ciphertext_sha256=NULL,
        plaintext_sha256=NULL,ciphertext_bytes=NULL,plaintext_bytes=NULL,
        row_count=NULL,filename=NULL,content_type=NULL,provider_request_id=NULL,
        row_revision=row_revision+1,updated_at=p_resolved_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN v_terminal_state;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_resolve_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,uuid,
    text,bytea,text,text,text,bytea,timestamptz,timestamptz,text,bytea,
    bytea,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_unknown_resolve_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,uuid,
    text,bytea,text,text,text,bytea,timestamptz,timestamptz,text,bytea,
    bytea,timestamptz,timestamptz
) TO forgecrud_runtime;
