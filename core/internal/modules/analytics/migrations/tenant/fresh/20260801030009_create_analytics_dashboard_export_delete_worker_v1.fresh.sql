-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_delete_snapshot_v1(
    p_export_id uuid, p_disposition text
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
BEGIN
    RETURN QUERY
    SELECT p_disposition,item.state,item.actor_id,item.export_id,
        item.dashboard_id,item.dashboard_version_id,item.token_version,
        item.authorization_version,item.authorization_clock,
        item.dashboard_manifest_sha256,item.pin_sha256,item.format,
        item.parameters_document,item.pinned_widgets,item.maximum_rows,
        item.maximum_plaintext_bytes,item.expires_at,item.attempt_generation,
        item.effect_started_at IS NOT NULL,COALESCE(item.object_key,''),
        COALESCE(item.object_version,''),COALESCE(item.key_id,''),
        COALESCE(item.wrapped_dek,''::bytea),COALESCE(item.wrap_nonce,''::bytea),
        COALESCE(item.header_sha256,''::bytea),
        COALESCE(item.ciphertext_sha256,''::bytea),
        COALESCE(item.plaintext_sha256,''::bytea),
        COALESCE(item.ciphertext_bytes,0),COALESCE(item.plaintext_bytes,0),
        COALESCE(item.row_count,0),COALESCE(item.content_type,''),
        COALESCE(item.filename,''),item.delete_actor_id,item.delete_job_id,
        item.delete_attempt_generation,COALESCE(item.delete_lease_owner,
            '00000000-0000-0000-0000-000000000000'::uuid)
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id;
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_delete_claim_v1(
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
DECLARE v_export analytics_dashboard_exports%ROWTYPE;
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
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.delete_job_id=p_delete_job_id
    FOR UPDATE;
    IF NOT FOUND OR v_export.delete_actor_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard export delete job was not found';
    END IF;
    IF v_export.state='deleted' OR
       v_export.state NOT IN ('running','deleting','cancelled') THEN
        RETURN QUERY SELECT * FROM analytics_dashboard_export_delete_snapshot_v1(
            p_export_id,'terminal');
        RETURN;
    END IF;
    IF v_export.state='running' AND
       v_export.cancellation_requested_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export delete was not requested';
    END IF;
    IF ((v_export.state='running' AND v_export.effect_started_at IS NOT NULL
         AND v_export.lease_until>p_observed_at) OR
        (v_export.state='deleting' AND
         v_export.delete_lease_until>p_observed_at)) THEN
        RAISE EXCEPTION USING ERRCODE='55P03',
            MESSAGE='analytics dashboard export delete lease is active';
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

CREATE FUNCTION analytics_dashboard_export_delete_complete_v1(
    p_scope_schema name,p_export_id uuid,p_delete_job_id uuid,
    p_lease_owner uuid,p_delete_generation bigint,p_object_version text,
    p_provider_request_id text,p_delete_response_sha256 bytea,
    p_absence_sha256 bytea,p_verified_at timestamptz,p_completed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_export analytics_dashboard_exports%ROWTYPE;
BEGIN
    IF p_delete_generation NOT BETWEEN 1 AND 3 OR
       (p_object_version IS NOT NULL AND
        pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512) OR
       (p_provider_request_id IS NOT NULL AND
        pg_catalog.octet_length(p_provider_request_id) NOT BETWEEN 1 AND 512) OR
       (p_delete_response_sha256 IS NOT NULL AND
        pg_catalog.octet_length(p_delete_response_sha256)<>32) OR
       (p_absence_sha256 IS NOT NULL AND
        pg_catalog.octet_length(p_absence_sha256)<>32) OR
       p_verified_at IS NULL OR NOT pg_catalog.isfinite(p_verified_at) OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       p_completed_at<p_verified_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export delete completion input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.state='deleting'
      AND item.delete_job_id=p_delete_job_id
      AND item.delete_lease_owner=p_lease_owner
      AND item.delete_attempt_generation=p_delete_generation FOR UPDATE;
    IF NOT FOUND OR (v_export.effect_started_at IS NOT NULL AND
       (p_provider_request_id IS NULL OR p_absence_sha256 IS NULL)) OR
       (v_export.object_version IS NOT NULL AND
        p_delete_response_sha256 IS NULL) THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='analytics dashboard export delete completion changed';
    END IF;
    UPDATE analytics_dashboard_export_attempt_receipts
    SET outcome='deleted',object_version=COALESCE(object_version,p_object_version),
        completed_at=p_completed_at
    WHERE export_id=p_export_id AND outcome='started';
    UPDATE analytics_dashboard_exports
    SET state='deleted',object_version=COALESCE(object_version,p_object_version),
        delete_provider_request_id=p_provider_request_id,
        delete_response_sha256=p_delete_response_sha256,
        delete_absence_sha256=p_absence_sha256,
        delete_verified_at=p_verified_at,deleted_at=p_completed_at,
        delete_lease_owner=NULL,delete_lease_until=NULL,
        row_revision=row_revision+1,updated_at=p_completed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_snapshot_v1(uuid,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_claim_v1(
    name,uuid,uuid,uuid,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_complete_v1(
    name,uuid,uuid,uuid,bigint,text,text,bytea,bytea,timestamptz,timestamptz
) FROM PUBLIC;
