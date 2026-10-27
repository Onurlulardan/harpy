-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_expiry_control_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,
    p_expiry_job_id uuid,p_observed_at timestamptz
)
RETURNS TABLE(state text,cleanup_required boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_job_id uuid;
    v_control_state text;
    v_control_changed_at timestamptz;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_update_at timestamptz;
    v_replay boolean;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_expiry_job_id) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export expiry control scope is invalid';
    END IF;
    SELECT item.job_id INTO v_job_id FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.expiry_job_id=p_expiry_job_id AND item.expires_at<=p_observed_at;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='Analytics dashboard export expiry candidate changed';
    END IF;
    SELECT control.control_state,control.changed_at
    INTO STRICT v_control_state,v_control_changed_at
    FROM platform.platform_async_analytics_dashboard_export_control_v1(
        p_tenant_id,p_export_id,v_job_id,p_observed_at
    ) AS control;
    SELECT item.* INTO STRICT v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=v_job_id AND item.expiry_job_id=p_expiry_job_id
      AND item.expires_at<=p_observed_at FOR UPDATE;
    v_replay:=v_export.delete_operation_id IS NOT DISTINCT FROM p_expiry_job_id AND
        v_export.delete_job_id IS NOT DISTINCT FROM p_expiry_job_id AND
        v_export.delete_actor_id IS NOT DISTINCT FROM v_export.actor_id AND
        v_export.delete_idempotency_key IS NOT DISTINCT FROM
            'analytics-expiry:'||p_export_id::text AND
        v_export.delete_request_sha256 IS NOT DISTINCT FROM pg_catalog.sha256(pg_catalog.convert_to(
            'analytics-expiry:'||p_export_id::text,'UTF8'));
    IF v_export.delete_operation_id IS NOT NULL AND NOT v_replay THEN
        state:=v_export.state;cleanup_required:=false;RETURN NEXT;RETURN;
    END IF;
    IF v_control_state='cancelled' THEN
        IF v_export.state NOT IN ('queued','expired') OR
           v_export.quota_state NOT IN ('none','released') THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export queued expiry differs';
        END IF;
        state:='expired';cleanup_required:=false;
    ELSIF v_control_state IN ('cancellation_requested','outcome_unknown') THEN
        IF v_export.state IS DISTINCT FROM 'running' OR
           v_export.quota_state IS DISTINCT FROM 'held' OR
           v_export.platform_attempt_id IS NULL OR
           (v_control_state='outcome_unknown' AND
            v_export.failure_code IS DISTINCT FROM 'ARTIFACT_OUTCOME_UNKNOWN') THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export active expiry differs';
        END IF;
        state:='expired';cleanup_required:=false;
    ELSIF v_control_state='succeeded' THEN
        IF v_export.state NOT IN ('ready','deleting','deleted') OR
           v_export.quota_state NOT IN ('used','released') THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export ready expiry differs';
        END IF;
        state:=v_export.state;
        cleanup_required:=v_export.state IN ('ready','deleting');
    ELSIF v_control_state IN ('dead','unknown_resolved') THEN
        IF v_export.state NOT IN ('failed','cancelled','expired','deleted') OR
           v_export.quota_state IS DISTINCT FROM 'released' THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export terminal expiry differs';
        END IF;
        state:=v_export.state;cleanup_required:=false;
    ELSE
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export expiry control is invalid';
    END IF;
    IF v_replay THEN RETURN NEXT;RETURN; END IF;
    v_update_at:=GREATEST(p_observed_at,v_control_changed_at,
        v_export.updated_at+interval '1 microsecond');
    UPDATE analytics_dashboard_exports
    SET state=CASE WHEN v_control_state='cancelled' THEN 'expired'
            ELSE analytics_dashboard_exports.state END,
        quota_state=CASE WHEN v_control_state='cancelled'
            THEN 'released' ELSE quota_state END,
        cancellation_requested_at=CASE WHEN v_control_state IN (
            'cancellation_requested','outcome_unknown') THEN
            COALESCE(cancellation_requested_at,p_observed_at)
            ELSE cancellation_requested_at END,
        delete_operation_id=p_expiry_job_id,delete_job_id=p_expiry_job_id,
        delete_actor_id=actor_id,
        delete_idempotency_key='analytics-expiry:'||p_export_id::text,
        delete_request_sha256=pg_catalog.sha256(pg_catalog.convert_to(
            'analytics-expiry:'||p_export_id::text,'UTF8')),
        row_revision=row_revision+1,updated_at=v_update_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN NEXT;
END;
$function$;

ALTER FUNCTION analytics_dashboard_export_delete_claim_v1(
    name,uuid,uuid,uuid,timestamptz
) RENAME TO analytics_dashboard_export_delete_claim_legacy_v1;

CREATE FUNCTION analytics_dashboard_export_delete_claim_v1(
    p_scope_schema name,p_export_id uuid,p_delete_job_id uuid,
    p_lease_owner uuid,p_observed_at timestamptz
)
RETURNS TABLE(
    disposition text,state text,actor_id uuid,export_id uuid,
    dashboard_id uuid,dashboard_version_id uuid,token_version bigint,
    authorization_version bigint,authorization_clock bigint,
    dashboard_manifest_sha256 bytea,pin_sha256 bytea,format text,
    parameters_document bytea,pinned_widgets jsonb,maximum_rows bigint,
    maximum_plaintext_bytes bigint,expires_at timestamptz,
    attempt_generation bigint,has_artifact boolean,object_key text,
    object_version text,key_id text,wrapped_dek bytea,wrap_nonce bytea,
    header_sha256 bytea,ciphertext_sha256 bytea,plaintext_sha256 bytea,
    ciphertext_bytes bigint,plaintext_bytes bigint,row_count bigint,
    content_type text,filename text,delete_actor_id uuid,delete_job_id uuid,
    delete_attempt_generation bigint,delete_lease_owner uuid
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_export analytics_dashboard_exports%ROWTYPE;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT item.* INTO STRICT v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id
      AND (item.delete_job_id=p_delete_job_id OR
           item.expiry_job_id=p_delete_job_id) FOR UPDATE;
    IF v_export.state='running' OR v_export.quota_state='held' OR
       v_export.failure_code='ARTIFACT_OUTCOME_UNKNOWN' OR EXISTS (
        SELECT 1 FROM analytics_dashboard_export_platform_attempts AS attempt
        WHERE attempt.export_id=p_export_id AND attempt.outcome='unknown'
          AND attempt.resolution_receipt_id IS NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export dedicated render cleanup owns artifact';
    END IF;
    IF p_delete_job_id=v_export.expiry_job_id AND
       v_export.delete_operation_id IS DISTINCT FROM p_delete_job_id THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export expiry control is required';
    END IF;
    RETURN QUERY SELECT *
    FROM analytics_dashboard_export_delete_claim_legacy_v1(
        p_scope_schema,p_export_id,p_delete_job_id,p_lease_owner,p_observed_at
    );
END;
$function$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_claim_legacy_v1(
    name,uuid,uuid,uuid,timestamptz
) FROM PUBLIC,harpy_runtime;
REVOKE ALL ON FUNCTION analytics_dashboard_export_expiry_control_v1(
    name,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_expiry_control_v1(
    name,uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_claim_v1(
    name,uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
