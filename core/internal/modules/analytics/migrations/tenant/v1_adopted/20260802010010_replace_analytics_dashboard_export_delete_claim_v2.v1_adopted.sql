-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_dashboard_export_delete_claim_v1(
    p_scope_schema name,p_export_id uuid,p_delete_job_id uuid,
    p_lease_owner uuid,p_observed_at timestamptz
)
RETURNS TABLE (
    disposition text,state text,actor_id uuid,export_id uuid,
    dashboard_id uuid,dashboard_version_id uuid,
    token_version bigint,authorization_version bigint,
    authorization_clock bigint,dashboard_manifest_sha256 bytea,
    pin_sha256 bytea,format text,parameters_document bytea,
    pinned_widgets jsonb,maximum_rows bigint,
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
DECLARE
    v_export analytics_dashboard_exports%ROWTYPE;
    v_tenant_id uuid;
    v_expiry boolean;
    v_has_artifact_evidence boolean;
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_delete_job_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export delete claim input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=pg_catalog.current_schema();
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=v_tenant_id
      AND (item.delete_job_id=p_delete_job_id OR
           item.expiry_job_id=p_delete_job_id)
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard export delete job was not found';
    END IF;
    v_expiry:=v_export.expiry_job_id=p_delete_job_id;
    IF v_expiry AND p_observed_at<v_export.expires_at THEN
        RAISE EXCEPTION USING ERRCODE='55P03',
            MESSAGE='analytics dashboard export expiry is not due';
    END IF;
    IF p_observed_at<v_export.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export delete clock is stale';
    END IF;
    IF v_expiry AND v_export.delete_operation_id IS NULL THEN
        UPDATE analytics_dashboard_exports AS target
        SET delete_operation_id=p_delete_job_id,delete_job_id=p_delete_job_id,
            delete_actor_id=target.actor_id,
            delete_idempotency_key='analytics-expiry:'||target.export_id::text,
            delete_request_sha256=pg_catalog.sha256(pg_catalog.convert_to(
                'analytics-expiry:'||target.export_id::text,'UTF8')),
            row_revision=target.row_revision+1,updated_at=p_observed_at
        WHERE target.export_id=p_export_id;
        SELECT item.* INTO STRICT v_export
        FROM analytics_dashboard_exports AS item
        WHERE item.export_id=p_export_id;
    ELSIF v_expiry AND v_export.delete_job_id<>p_delete_job_id THEN
        RETURN QUERY SELECT * FROM analytics_dashboard_export_delete_snapshot_v1(
            p_export_id,'terminal');
        RETURN;
    END IF;
    IF v_export.delete_actor_id IS NULL OR
       v_export.delete_job_id IS DISTINCT FROM p_delete_job_id THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard export delete job was not found';
    END IF;
    IF v_export.state='deleted' THEN
        RETURN QUERY SELECT * FROM analytics_dashboard_export_delete_snapshot_v1(
            p_export_id,'terminal');
        RETURN;
    END IF;
    IF ((v_export.state='running' AND v_export.lease_until>p_observed_at) OR
        (v_export.state='deleting' AND
         v_export.delete_lease_until>p_observed_at)) THEN
        RAISE EXCEPTION USING ERRCODE='55P03',
            MESSAGE='analytics dashboard export delete lease is active';
    END IF;
    v_has_artifact_evidence:=v_export.effect_started_at IS NOT NULL OR
        v_export.object_version IS NOT NULL OR v_export.quota_state='used' OR
        v_export.state IN ('ready','deleting');
    IF v_expiry AND NOT v_has_artifact_evidence THEN
        IF v_export.quota_state='held' THEN
            UPDATE platform.artifact_quota_counters
            SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
                revision=revision+1,updated_at=p_observed_at
            WHERE tenant_id=v_tenant_id
              AND reserved_ciphertext_bytes>=2147483648;
            IF NOT FOUND THEN
                RAISE EXCEPTION USING ERRCODE='P6105',
                    MESSAGE='analytics dashboard export expiry quota changed';
            END IF;
        END IF;
        UPDATE analytics_dashboard_exports AS target
        SET state=CASE WHEN target.state IN ('queued','running') THEN 'expired'
                       ELSE target.state END,
            quota_state=CASE WHEN target.quota_state IN ('none','held')
                             THEN 'released' ELSE target.quota_state END,
            lease_owner=NULL,lease_until=NULL,row_revision=target.row_revision+1,
            updated_at=p_observed_at
        WHERE target.export_id=p_export_id;
        RETURN QUERY SELECT * FROM analytics_dashboard_export_delete_snapshot_v1(
            p_export_id,'terminal');
        RETURN;
    END IF;
    IF NOT v_has_artifact_evidence THEN
        RETURN QUERY SELECT * FROM analytics_dashboard_export_delete_snapshot_v1(
            p_export_id,'terminal');
        RETURN;
    END IF;
    UPDATE analytics_dashboard_exports
    SET state='deleting',delete_attempt_generation=LEAST(
            analytics_dashboard_exports.delete_attempt_generation+1,3),
        delete_lease_owner=p_lease_owner,
        delete_lease_until=p_observed_at+INTERVAL '60 seconds',
        lease_owner=NULL,lease_until=NULL,row_revision=row_revision+1,
        updated_at=p_observed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN QUERY SELECT * FROM analytics_dashboard_export_delete_snapshot_v1(
        p_export_id,'execute');
END;
$function$;
-- +goose StatementEnd
