-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_dashboard_export_delete_complete_v1(
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
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_delete_job_id) IS DISTINCT FROM 7 OR
       p_lease_owner IS NULL OR
       p_lease_owner='00000000-0000-0000-0000-000000000000'::uuid OR
       p_delete_generation NOT BETWEEN 1 AND 3 OR
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
    IF v_export.quota_state='held' THEN
        UPDATE platform.artifact_quota_counters
        SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
            revision=revision+1,updated_at=p_completed_at
        WHERE tenant_id=v_export.tenant_id
          AND reserved_ciphertext_bytes>=2147483648;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6105',
                MESSAGE='analytics dashboard export reserved quota changed';
        END IF;
    ELSIF v_export.quota_state='used' THEN
        UPDATE platform.artifact_quota_counters
        SET used_ciphertext_bytes=used_ciphertext_bytes-v_export.ciphertext_bytes,
            revision=revision+1,updated_at=p_completed_at
        WHERE tenant_id=v_export.tenant_id
          AND used_ciphertext_bytes>=v_export.ciphertext_bytes;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6105',
                MESSAGE='analytics dashboard export used quota changed';
        END IF;
    ELSIF v_export.quota_state NOT IN ('none','released') THEN
        RAISE EXCEPTION USING ERRCODE='P6105',
            MESSAGE='analytics dashboard export delete quota is invalid';
    END IF;
    UPDATE analytics_dashboard_export_attempt_receipts
    SET outcome='deleted',object_version=COALESCE(object_version,p_object_version),
        completed_at=p_completed_at
    WHERE export_id=p_export_id AND outcome='started';
    UPDATE analytics_dashboard_exports
    SET state='deleted',quota_state='released',
        object_version=COALESCE(object_version,p_object_version),
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

