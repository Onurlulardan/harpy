-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_platform_succeed_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,p_effect_receipt_id uuid,
    p_effect_result_sha256 bytea,p_object_version text,
    p_provider_request_id text,p_uploaded_at timestamptz,
    p_completed_at timestamptz
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_attempt analytics_dashboard_export_platform_attempts%ROWTYPE;
    v_exact boolean;
    v_ciphertext_bytes bigint;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation IS NULL OR p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_effect_result_sha256) IS DISTINCT FROM 32 OR
       COALESCE(pg_catalog.octet_length(p_object_version),0) NOT BETWEEN 1 AND 512 OR
       COALESCE(pg_catalog.octet_length(p_provider_request_id),0) NOT BETWEEN 1 AND 512 OR
       p_uploaded_at IS NULL OR NOT pg_catalog.isfinite(p_uploaded_at) OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       p_completed_at<p_uploaded_at THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export success scope is invalid';
    END IF;
    SELECT item.ciphertext_bytes INTO v_ciphertext_bytes
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id;
    IF NOT FOUND OR v_ciphertext_bytes IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export success attempt is stale';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_succeed_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_lease_owner,p_effect_id,p_effect_request_sha256,p_effect_receipt_id,
        p_effect_result_sha256,p_provider_request_id,
        v_ciphertext_bytes,p_completed_at);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.job_id=p_job_id FOR UPDATE;
    SELECT attempt.* INTO v_attempt
    FROM analytics_dashboard_export_platform_attempts AS attempt
    WHERE attempt.export_id=p_export_id
      AND attempt.attempt_generation=p_attempt_generation
      AND attempt.platform_attempt_id=p_attempt_id
      AND attempt.job_id=p_job_id AND attempt.lease_owner=p_lease_owner
    FOR UPDATE;
    IF v_export.export_id IS NULL OR v_attempt.export_id IS NULL OR
       v_export.ciphertext_bytes IS DISTINCT FROM v_ciphertext_bytes THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export success attempt is stale';
    END IF;
    IF v_attempt.outcome='ready' THEN
        SELECT receipt.platform_attempt_id IS NOT DISTINCT FROM p_attempt_id AND
               receipt.platform_effect_id IS NOT DISTINCT FROM p_effect_id AND
               receipt.platform_effect_request_sha256 IS NOT DISTINCT FROM
                   p_effect_request_sha256 AND
               receipt.platform_effect_receipt_id IS NOT DISTINCT FROM
                   p_effect_receipt_id AND
               receipt.platform_effect_result_sha256 IS NOT DISTINCT FROM
                   p_effect_result_sha256 AND
               receipt.object_version IS NOT DISTINCT FROM p_object_version AND
               receipt.provider_request_id IS NOT DISTINCT FROM
                   p_provider_request_id AND
               receipt.uploaded_at IS NOT DISTINCT FROM p_uploaded_at AND
               receipt.completed_at IS NOT DISTINCT FROM p_completed_at INTO v_exact
        FROM analytics_dashboard_export_attempt_receipts AS receipt
        WHERE receipt.export_id=p_export_id
          AND receipt.attempt_generation=p_attempt_generation FOR UPDATE;
        IF v_export.state IS DISTINCT FROM 'ready' OR
           v_export.quota_state IS DISTINCT FROM 'used' OR
           v_export.object_version IS DISTINCT FROM p_object_version OR
           v_export.provider_request_id IS DISTINCT FROM p_provider_request_id OR
           v_export.ready_at IS DISTINCT FROM p_completed_at OR
           v_export.updated_at IS DISTINCT FROM p_completed_at OR
           v_export.lease_owner IS NOT NULL OR v_export.lease_until IS NOT NULL OR
           v_export.platform_attempt_id IS NOT NULL OR
           v_export.platform_effect_id IS NOT NULL OR
           v_export.platform_effect_request_sha256 IS NOT NULL OR
           v_attempt.platform_effect_id IS DISTINCT FROM p_effect_id OR
           v_attempt.platform_effect_request_sha256 IS DISTINCT FROM
               p_effect_request_sha256 OR
           v_attempt.platform_effect_receipt_id IS DISTINCT FROM
               p_effect_receipt_id OR
           v_attempt.platform_effect_result_sha256 IS DISTINCT FROM
               p_effect_result_sha256 OR
           v_attempt.uploaded_at IS DISTINCT FROM p_uploaded_at OR
           v_attempt.finished_at IS DISTINCT FROM p_completed_at OR
           NOT COALESCE(v_exact,false) THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export success replay differs';
        END IF;
        RETURN 'ready';
    END IF;
    IF v_attempt.outcome IS DISTINCT FROM 'running' OR
       v_export.state IS DISTINCT FROM 'running' OR
       v_export.quota_state IS DISTINCT FROM 'held' OR
       v_export.platform_attempt_id IS DISTINCT FROM p_attempt_id OR
       v_export.platform_effect_id IS DISTINCT FROM p_effect_id OR
       v_export.platform_effect_request_sha256 IS DISTINCT FROM
           p_effect_request_sha256 OR
       v_export.attempt_generation IS DISTINCT FROM p_attempt_generation OR
       v_export.lease_owner IS DISTINCT FROM p_lease_owner OR
       v_export.effect_started_at IS NULL OR
       p_uploaded_at<v_export.effect_started_at OR
       v_export.cancellation_requested_at IS NOT NULL OR
       v_export.expires_at<=p_completed_at OR
       p_completed_at<=v_export.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export success lease is stale';
    END IF;
    UPDATE analytics_dashboard_export_attempt_receipts
    SET outcome='ready',object_version=p_object_version,
        provider_request_id=p_provider_request_id,uploaded_at=p_uploaded_at,
        completed_at=p_completed_at,
        platform_effect_receipt_id=p_effect_receipt_id,
        platform_effect_result_sha256=p_effect_result_sha256
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation AND outcome='started';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export success receipt changed';
    END IF;
    UPDATE analytics_dashboard_export_platform_attempts
    SET outcome='ready',platform_effect_receipt_id=p_effect_receipt_id,
        platform_effect_result_sha256=p_effect_result_sha256,
        uploaded_at=p_uploaded_at,finished_at=p_completed_at
    WHERE export_id=p_export_id
      AND attempt_generation=p_attempt_generation AND outcome='running';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export success attempt changed';
    END IF;
    UPDATE analytics_dashboard_exports
    SET state='ready',quota_state='used',object_version=p_object_version,
        provider_request_id=p_provider_request_id,ready_at=p_completed_at,
        failure_code=NULL,lease_owner=NULL,lease_until=NULL,
        platform_attempt_id=NULL,platform_effect_id=NULL,
        platform_effect_request_sha256=NULL,row_revision=row_revision+1,
        updated_at=p_completed_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN 'ready';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_succeed_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,bytea,text,text,
    timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_succeed_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,bytea,text,text,
    timestamptz,timestamptz) TO harpy_runtime;
